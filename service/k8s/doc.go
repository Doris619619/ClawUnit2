// Package k8s 是 ClawUnit 与 Kubernetes 交互的根包。
//
// 它本身只暴露最基础的东西：全局 Client 单例、初始化逻辑、命名工具
// 函数。具体的资源 CRUD 都拆到了子包里：
//
//	pod/            Pod 创建、查询、删除、状态读取
//	configmap/      openclaw.json 渲染、ConfigMap CRUD、插件 patch
//	networkpolicy/  egress 网络策略
//	pvc/            实例数据 PVC 的 CRUD（含孤立 PVC 列出与回收）
//	namespace/      用户级 namespace 创建
//	exec/           在 Pod 内执行命令（同步 + 流式）
//	portforward/    out-of-cluster 模式的 SPDY port-forward 连接缓存
//	cleanup/        按 label selector 批量清理实例资源
//
// 这种拆分是因为 service 层的代码风格约定是"一个文件只放一个公共
// 函数"，整包内函数太多就会爆出几十个文件，可读性反而下降。子包让
// 函数名也变短：pod.Create 替代 k8s.CreatePod，namespace.Ensure 替代
// k8s.EnsureNamespace。子包都依赖父包的 Client 单例 —— 通过 GetClient()
// 拿到，不要在子包里再做配置。
//
// # K8s 资源模型
//
// ClawUnit 把每个用户的资源放在独立 namespace 里：
//
//	namespace = "{base}-user-{upn_hash8}"   // base 来自 k8s.namespace 配置
//
// 每个实例对应一组 K8s 资源：
//
//	Pod              {pod}/clawunit-{id}-{name}                运行 OpenClaw Gateway
//	Service          {svc}/clawunit-{id}-{name}                ClusterIP，暴露 18789
//	ConfigMap        {cm}/clawunit-{id}-config                 openclaw.json 配置
//	NetworkPolicy    {netpol}/clawunit-{id}-{name}-netpol      egress 限制
//	用户数据 PVC      clawunit-{name}-{upn_hash}                持久化 OpenClaw 运行数据
//
// 整组资源都打了 label `app=clawunit, instance-id=<id>, managed-by=clawunit`，
// 删除时按 label selector 一键清理（见 cleanup.DeleteAllResources）。
//
// # 共享 PVC
//
// 系统技能和 Playwright 浏览器是跨用户共享的，但 K8s NetworkPolicy 不
// 支持跨 namespace 引用 PVC，所以管理员需要在每个用户 namespace 里
// 预创建一份 SystemSkillsPVC 和 PlaywrightPVC。Pod 启动时只读挂载这两个
// PVC，避免用户篡改。
//
// 用户数据 PVC 是每个实例一个，跨实例不共享 —— 这样删除一个实例不会
// 影响其他实例的数据。删除实例时默认不删除 PVC，前端通过
// ListOrphanPVCs/DeletePVC 让用户自己管理。
//
// # 命名规范
//
// 所有 K8s 名称都经过 sanitizeK8sName 规整，保证符合 DNS-1123 label 规则
// （小写字母数字和 -，长度 ≤ 63）。UPN 因为可能含特殊字符，先取 SHA256
// 前 4 字节（8 个 hex 字符）作为短标识 —— 见 UpnHash。
//
// # in-cluster 与 out-of-cluster
//
// Initialize 支持三种模式：
//
//   - "incluster"：用 ServiceAccount token，部署到 K8s 集群内运行
//   - "outofcluster"：用 kubeconfig 文件，本地开发用
//   - "auto"（默认）：先试 incluster，失败回退到 outofcluster
//
// out-of-cluster 模式下连 Pod 不能直连 Pod IP，要走 portforward 子包
// 建立 SPDY tunnel，详见 internal/service/k8s/portforward。
package k8s
