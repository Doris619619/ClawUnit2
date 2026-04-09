// Package v1 是管理员 API 的 request/response 定义。
//
// 路由列表：
//
//	AdminListInstancesReq  GET  /api/admin/v1/instances/list   全局实例列表（可按 ownerUpn 过滤）
//	AdminUpdateQuotaReq    POST /api/admin/v1/quotas/update    设置用户的最大实例数
//	SyncStatusReq          GET  /api/admin/v1/sync/status      获取 K8s 同步服务运行状态
//	ForceSyncReq           POST /api/admin/v1/sync/force       立即触发一次 K8s 同步
//
// 这些接口都需要 (clawunit, admin) 权限。Admin 与普通用户的区分在
// internal/middlewares.RequireAdmin 完成，UniAuth 是事实来源。
//
// AdminInstanceItem 与 instances/v1.InstanceItem 的区别：管理员视角
// 多展示 OwnerUpn（所属用户），少展示一些用户视角的 derived 字段。
package v1
