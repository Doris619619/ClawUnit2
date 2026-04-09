// Package instances 实现实例管理 HTTP API。
//
// 这是 api/instances/v1 中 IInstancesV1 interface 的具体实现。每个 API
// 方法独占一个文件（instances_v1_create.go、instances_v1_get_list.go 等），
// 这是项目代码风格约定 —— 详见 CLAUDE.md 的 "File organization" 一节。
//
// # 文件结构
//
//	instances.go            // gf gen ctrl 生成的占位文件，用于放包级 helper
//	instances_new.go        // gf gen ctrl 生成的 NewV1 工厂
//	instances_v1_*.go       // 每个 API 方法一个文件，文件名固定，不能改
//
// # 控制层职责
//
// Controller 是 HTTP 边界层，职责包括：
//
//   - 从 ctx 提取 ownerUpn（用 middlewares.OwnerFromCtx）
//   - 把 v1 request 字段转成 service 层的入参
//   - 调用 dao 或 service 层做实际工作
//   - 把结果包成 v1 response 返回
//   - 失败时返回 gerror.Wrapf(err, "...") 或 gerror.NewCodef(gcode.CodeNotFound, "...")
//
// 控制层不应该做：业务逻辑、K8s 资源操作、跨实例的状态机管理 ——
// 这些都在 internal/service/lifecycle 里。
//
// # 直接调 dao 是允许的
//
// 简单的查询（例如 GetList、GetOne）controller 直接 dao.Instances.Ctx(ctx)...
// 即可，没必要再加一层 service 包装。需要 K8s 操作或跨表事务时才走
// internal/service/lifecycle。
package instances
