// Package admin 实现管理员 HTTP API。
//
// 实现 api/admin/v1 中的 IAdminV1 interface。所有方法都需要 admin 权限，
// 通过路由组挂载 middlewares.RequireAdmin 中间件统一校验，controller
// 内部不再重复检查。
//
// 文件结构遵循 gf gen ctrl 约定：每个 API 方法一个文件
// （admin_v1_admin_list_instances.go 等），不要把多个方法塞到同一个文件里。
package admin
