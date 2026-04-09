// Package sync 把 K8s Pod 实时状态同步到 instances 表。
//
// 公开函数列表：
//
//	Start         启动后台同步 goroutine（应用启动时调用一次）
//	ForceSync     立即触发一次同步（管理员 API 用）
//	LastRunTime   返回上次同步的 RFC3339 时间戳，admin 仪表盘展示用
//
// # 为什么需要这个服务
//
// lifecycle 包负责创建/删除 Pod，但 Pod 从 Pending → Running 是 K8s
// 异步推进的，ClawUnit 没法在 lifecycle.Create 里同步等待 Ready
// （那样接口会阻塞十几秒到几分钟）。所以方案是：
//
//  1. lifecycle.Create 写入 status="creating" 立即返回
//  2. sync 服务每 5 秒轮询所有 creating/running/error 实例的 Pod
//  3. 根据 Pod Phase + Ready Condition 推进状态
//
// 这样所有"被动状态变化"（Pod 崩溃、节点宕机、用户在 K8s 里手工删
// Pod）都能被及时反映到 DB。
//
// # 状态映射规则
//
//	K8s Pod Phase + Conditions  →  ClawUnit Status
//	-------------------------------------------------
//	Running + Ready=True        →  running
//	Running + Ready=False       →  保持当前状态（仍在 startup 中）
//	Pending                     →  保持当前状态（保持 creating）
//	Succeeded                   →  stopped
//	Failed                      →  error
//	Unknown                     →  保持当前状态
//	Pod 不存在 + 当前 running    →  stopped（被外部删了）
//	Pod 不存在 + 当前 creating   →  error（创建中却消失了，肯定出问题）
//
// # Pod IP 同步
//
// 实例进入 running 状态后，sync 会把 Pod IP 写入 pod_ip 列。这是
// proxy.WsChat 在 in-cluster 模式下直连 Pod 的依据 —— 不走 K8s
// Service，因为 ClusterIP 多一跳延迟，对 WS 桥来说没必要。
//
// # 启动失败的处理
//
// Start 在第一步就调 k8s.Initialize，如果初始化失败（比如 kubeconfig
// 不存在、API server 不可达），会写 error 日志后 *直接 return*，
// 同步服务停掉而 ClawUnit HTTP 服务不会挂。这样配置错误时还能通过
// HTTP API 修复 —— 但所有依赖 K8s 的接口都会失败。
//
// # 不要在 sync 里改业务逻辑
//
// 本包只负责 Pod Phase → DB status 的镜像，不能引入"如果连续 N 次
// 失败就自动 Restart"之类的行为。那种自愈逻辑要写在专门的 reconciler
// 里，并且需要明确的去抖动和速率限制。
package sync
