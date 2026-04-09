// Package middlewares 提供 ClawUnit HTTP 服务的中间件链。
//
// 中间件分两类：
//
//   - 响应包装：UniResMiddleware 把所有 controller 返回值包成统一的
//     {success, code, message, data} JSON 结构，与 open-platform 保持一致。
//     流式响应（SSE / octet-stream / multipart）会自动跳过包装。
//   - 身份注入与鉴权：InjectIdentity 从 X-User-ID 请求头取出 UPN，
//     调用 UniAuth 验证 (clawunit, access) 权限，然后写入 ctx；RequireAdmin
//     在此基础上再校验 (clawunit, admin) 权限。
//
// # 中间件顺序
//
// 必须按照 UniRes → InjectIdentity → RequireAdmin（如有）的顺序挂载。
// 顺序反了会出现两类 bug：
//
//   - InjectIdentity 在 UniRes 之前：身份失败时返回的 401 不会被包装
//     成统一格式，前端拿到的是裸 JSON。
//   - RequireAdmin 在 InjectIdentity 之前：ctx 里没有 ownerUpn，
//     RequireAdmin 一定 401。
//
// # 从 ctx 里取用户身份
//
// InjectIdentity 把 UPN 存在 ctx 的 "ownerUpn" key 下。controller 不要
// 直接用类型断言读取（forcetypeassert lint 会报错），而是用 OwnerFromCtx：
//
//	func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (*v1.CreateRes, error) {
//	    ownerUpn := middlewares.OwnerFromCtx(ctx)
//	    // ...
//	}
//
// 经过 InjectIdentity 中间件的请求保证 ownerUpn 非空。
//
// # 流式响应识别
//
// UniResMiddleware 通过检查 Content-Type 决定是否跳过包装，识别这三种 MIME：
//
//   - text/event-stream（SSE，渠道扫码登录用）
//   - application/octet-stream（二进制流，Transfer.Export 的 tar.gz、
//     图片代理 LatestMedia 的 PNG/JPEG）
//   - multipart/x-mixed-replace（保留给将来的 MJPEG）
//
// 如果你新增了流式接口，记得让 handler 在写第一字节前就 SetContentType，
// 否则中间件会以为是普通 JSON 响应去包装它。
package middlewares
