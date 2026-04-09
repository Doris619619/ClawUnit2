// Package cleanup 按 label selector 批量清理实例的 K8s 资源。
//
// 公开函数列表：
//
//	DeleteAllResources   删除一个实例的 Pod / Service / NetworkPolicy
//	WaitForPodDeletion   阻塞等待匹配的 Pod 全部消失
//
// # 删除范围
//
// DeleteAllResources 通过 selector
//
//	instance-id={id},managed-by=clawunit
//
// 列出并删除三类资源：Pod、Service、NetworkPolicy。*不删除*：
//
//   - ConfigMap：调用方应单独调 configmap.Delete（API 路径不同，不能放进同一个 selector 删除流程）
//   - PVC：用户数据保留策略，由 lifecycle.Delete 决定是否额外调 pvc.DeleteInstance
//   - Namespace：跨实例共享，永远不删
//
// 这是 lifecycle.Delete 的内部步骤之一 —— 业务层调用方不应该直接用
// 本包，应该用 lifecycle.Delete 走完整流程。
//
// # 错误处理
//
// 任何子操作失败都只 warning，不中断后续清理。原因：删除是"尽力而为"
// 的操作，部分失败也要继续往下走，避免半清理状态。失败的资源会留下
// 日志，运维可以手动处理。
//
// # 为什么有 //nolint:dupl
//
// Pod / Service / NetworkPolicy 的清理代码长得几乎一样（list →
// for-loop → delete），但 K8s clientset 的 API 路径不同
// （CoreV1().Pods vs NetworkingV1().NetworkPolicies），泛型也救不了
// 因为返回类型完全不同。dupl linter 误报，加 //nolint:dupl 豁免。
package cleanup
