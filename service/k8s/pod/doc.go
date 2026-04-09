// Package pod 管理 OpenClaw 实例的 Pod 资源。
//
// 公开函数列表：
//
//	Create         创建实例 Pod，返回创建后的 *corev1.Pod
//	Get            按 instance-id label 查找 Pod
//	GetStatus      查询 Pod 的 corev1.PodStatus（含 Phase / Conditions / PodIP）
//	Delete         按 instance-id label 删除 Pod（不存在视为成功）
//	DeleteAndWait  Delete 之后阻塞最多 30 秒等 Pod 实际消失
//
// # Pod 结构
//
// 每个 OpenClaw 实例都是单 Pod 单容器（gateway），外加一个 init container：
//
//	init-config (busybox)         首次启动时把 ConfigMap 里的 openclaw.json
//	                              复制到 PVC，确保后续运行时改动可持久化
//	gateway (openclaw 镜像)        实际跑 OpenClaw Gateway 的容器
//
// 容器以 root 启动，因为需要 apt 安装 Chromium 系统依赖，启动命令长这样：
//
//	playwright install-deps chromium > /dev/null 2>&1
//	exec runuser -u node -- node /app/dist/index.js gateway run
//
// 装完依赖后立刻 runuser 切换到 node 用户运行 gateway 进程，最大限度
// 缩小 root 时间窗口。AllowPrivilegeEscalation = false 在切完用户后
// 防止再提权回去。
//
// # 健康检查
//
// 三个 probe 都用 exec（不用 httpGet），因为 OpenClaw Gateway bind 到 lan
// 而不是 0.0.0.0，从 K8s probe 容器里访问 Pod IP 会被拒。改成在容器
// 内跑 node -e 'require("http").get(...)' 直接打 127.0.0.1 即可：
//
//	StartupProbe   /healthz   FailureThreshold=120 PeriodSeconds=5  (10 分钟启动窗口)
//	ReadinessProbe /readyz    PeriodSeconds=10
//	LivenessProbe  /healthz   PeriodSeconds=30
//
// 启动窗口长是因为 Chromium 依赖安装可能很慢（500MB 的 apt 包），
// 加上 OpenClaw 自检会拉模型 metadata。
//
// # Volume 挂载
//
//	openclaw-config    ConfigMap   /config              init container 用，复制完即弃
//	instance-data      PVC RW      /home/node/.openclaw 用户数据，整目录持久化
//	system-skills      PVC RO      /skills/system       管理员预置技能
//	playwright-browsers PVC RO     /home/node/.cache/ms-playwright  浏览器二进制
//	tmp-volume         EmptyDir    /tmp                 临时文件
//	dshm               EmptyDir(memory) /dev/shm        Chromium 共享内存
//
// dshm 用 memory-backed EmptyDir 是为了解决 Chrome 在默认 64MB /dev/shm
// 上跑出 timeout 的问题。
//
// # GPU 调度
//
// GPUEnabled + GPUCount > 0 时会给容器 Limits/Requests 加上
// nvidia.com/gpu，K8s 的 nvidia device plugin 自动接管调度。
//
// # AlreadyExists 处理
//
// Create 如果撞上 AlreadyExists（通常是上次创建残留没清干净），会
// 自动删除旧 Pod、等 5 秒、重新创建一次。这是相对粗暴的策略，但
// 因为 ClawUnit 是单实例 Pod，没有 deployment rolling update 的需求，
// 直接覆盖最简单也最可靠。
package pod
