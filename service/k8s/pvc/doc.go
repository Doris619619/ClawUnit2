// Package pvc 管理实例数据 PVC 的生命周期。
//
// 公开函数列表：
//
//	GetInstanceName   生成实例 PVC 的标准名称
//	EnsureInstance    创建实例 PVC（或验证复用历史 PVC）
//	DeleteInstance    删除实例 PVC（按名字）
//	ListOrphan        列出用户 namespace 下未被任何活跃实例引用的 PVC
//
// # PVC 的所有权
//
// 用户数据 PVC 是 *每实例一个*，跨实例不共享。命名规则：
//
//	clawunit-{instance_name}-{upn_hash8}
//
// PVC label 上记录 owner-upn（hash 后），EnsureInstance 复用历史 PVC 时
// 会校验 label 防止越权挂载别人的数据。
//
// # 不删除 PVC 的策略
//
// 删除实例时 *默认不删除 PVC*，保留用户数据。前端通过 ListOrphan 让
// 用户看到所有未被引用的 PVC，可以选择：
//
//   - 创建新实例时绑定历史 PVC（CreateReq.ExistingPvc 字段）
//   - 调 DeletePVC 显式回收存储
//
// 这样做的好处：误删实例的成本很低（重新创建一个绑定旧 PVC 即可），
// 代价是空 PVC 会持续占用存储配额。
//
// # 异步绑定监控
//
// PVC 创建后会启动一个 background goroutine 监控 .Status.Phase，
// 等到 ClaimBound 写一条日志，超时（30s）也只是 warning 不报错 ——
// 因为 Pod 启动会自动等 PVC 绑定，我们这条 goroutine 只是给运维看
// 状态用。注意它用 context.Background()（不是请求 ctx），避免请求
// 结束后 goroutine 提前被取消，linter 通过 //nolint:gosec,contextcheck
// 标注豁免。
//
// # 系统技能 PVC 和 Playwright PVC 不在这里
//
// 这两个跨用户共享的 PVC 是管理员手工创建的，不归本包管理。本包
// 只管 *实例数据* PVC（每实例一个）。
package pvc
