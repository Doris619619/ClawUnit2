// Package lifecycle 编排 OpenClaw 实例的完整生命周期。
//
// 公开函数列表：
//
//	Create     配额检查 → DB 记录 → API key → namespace → ConfigMap → NetworkPolicy → PVC → Pod
//	Start      从 stopped/error 状态恢复实例（重新创建 Pod，复用已有 ConfigMap/PVC）
//	Stop       删除 Pod，保留所有持久资源（ConfigMap/PVC/NetworkPolicy 不动）
//	Restart    Stop 然后 Start
//	Delete     关闭 port-forward → 等 Pod 终止 → 删 NetworkPolicy/ConfigMap → 吊销 API Key → 删 DB 记录
//	GetStatus  返回实例的 DB status + Pod 实时 Phase + Ready 条件
//
// 这是 internal/service/k8s 子包之上的编排层 —— k8s 子包只管单个
// K8s 资源 CRUD，本包负责"实例 = 一组资源"的完整流程，包括：
//
//   - 跨 K8s 资源的依赖顺序（namespace 必须先于 ConfigMap，ConfigMap 必须先于 Pod）
//   - 任意步骤失败的回滚
//   - DB 记录与 K8s 资源的一致性
//   - UniAuth API Key 的发放/恢复/吊销
//
// # 状态机
//
// 实例的 status 字段在 instances 表里，状态值与转换：
//
//	creating  ←─ Create() 写入的初始状态
//	   │
//	   ↓ (sync 服务监听到 Pod Ready)
//	running   ←─ 正常工作状态
//	   │
//	   ├─ Stop() ──→ stopped
//	   │              │
//	   │              └─ Start() ──→ creating ──→ running
//	   │
//	   └─ (Pod 异常崩溃) ──→ error
//	                          │
//	                          └─ Start() ──→ creating ──→ running
//
// status = "deleting" 是 Delete() 进行中的瞬态，正常情况下完成后整条
// 记录会被删除。如果 Delete 中途崩溃，会留下 deleting 记录给人工
// 介入。
//
// # 状态推进的来源
//
// lifecycle 包只把 status 写成 creating（Create / Start）或 stopped
// （Stop）。把 creating → running 的推进交给 internal/service/sync —
// 它每 5 秒轮询活跃实例的 Pod Phase，根据 Ready 条件更新状态。
// 这样状态机有单一来源，避免并发写冲突。
//
// # Create 的回滚顺序
//
// Create 是 11 个步骤的长流程，任何一步失败都要回滚之前已经创建的
// 资源 + DB 记录 + API Key（如果是 auto 模式）。回滚顺序是反向的，
// 用 if err != nil 链 + revokeAndCleanup helper 实现。这套代码乍看
// 重复但逐层析构是必要的 —— defer 解决不了，因为部分资源（PVC）
// 即使创建成功也不应该回滚（用户数据保留策略）。
//
// # API Key 的双模式
//
//   - manual 模式：用户提供 ApiKey + BaseUrl，直接当环境变量注入。
//   - auto 模式：调 uniauth.CreateApiKey 拿 raw key + hash，
//     raw key 注入容器环境变量，hash 存 DB 用于后续吊销。
//     注意：raw key *没有持久化*，Stop → Start 时需要 ensureApiKey
//     重新创建一份新 key（旧 key 同时被吊销）。
//
// # Controller 不要绕过 lifecycle 直接编排实例
//
// 实例的"创建/启动/停止/重启/删除"动作必须走本包的对应函数 ——
// 不要在 controller 里手工调用 namespace.Ensure / pod.Create / cleanup
// 等子包拼出一个 ad-hoc 的实例。那样会绕过配额检查、回滚和 DB 状态机。
//
// 这条约束由 .golangci.yml 的 `controller_must_not_depend_on_lifecycle_k8s`
// depguard 规则强制：controller 包禁止 import 这几个 K8s 子包：
//
//	internal/service/k8s/pod
//	internal/service/k8s/namespace
//	internal/service/k8s/networkpolicy
//	internal/service/k8s/cleanup
//
// 仍然允许 controller 直接 import 的 K8s 子包：
//
//	exec        渠道插件管理需要在 Pod 内执行命令
//	configmap   配置热更新（UpdateConfig API）需要直接 patch ConfigMap
//	pvc         孤立 PVC 的 list/delete 是用户级数据管理，与实例生命周期无关
//	k8s（根包） 拿 Client 单例做名称计算等只读操作
//
// 如果 controller 需要查 Pod 的实时状态（GetStatus 之类的只读操作），
// 应该通过 lifecycle 暴露的入口（例如 lifecycle.GetStatus），而不是
// 直接 import k8s/pod 子包。
package lifecycle
