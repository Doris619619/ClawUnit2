// Package channels 定义渠道插件管理 API 的 controller interface。
//
// 由 `gf gen ctrl` 从 api/channels/v1/channels.go 自动生成。
//
// 注意 LoginReq 不在这个 interface 里 —— 它是 SSE 流式接口，必须用
// 原生 ghttp.Request handler 注册（见 internal/cmd/cmd.go 的
// channels.LoginSSE 路由），无法通过 controller.Bind 自动绑定。
package channels
