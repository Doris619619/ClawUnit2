// Package v1 是渠道插件管理 API 的 request/response 定义。
//
// 路由列表：
//
//	InstallReq          POST /api/channels/v1/install          安装渠道插件
//	UninstallReq        POST /api/channels/v1/uninstall        卸载渠道插件
//	RestartGatewayReq   POST /api/channels/v1/restart-gateway  重启 Gateway 进程加载新插件
//	StatusReq           GET  /api/channels/v1/status           查询渠道连接状态
//	LoginReq            POST /api/channels/v1/login            扫码登录（SSE 流式，不在 interface 里）
//
// # 渠道插件的物理位置
//
// 插件包安装到用户数据 PVC 的 extensions/ 目录，跨实例重启不会丢失。
// 安装完成后通过 internal/service/k8s/configmap.PatchPlugin 更新
// ConfigMap 的 plugins.allow 和 plugins.entries，再调 RestartGateway
// 让 OpenClaw 进程加载新插件 —— 不会重启容器，所以已建立的 WebSocket
// session 不受影响。
//
// # 为什么 LoginReq 不在 IChannelsV1 里
//
// 扫码登录返回的是二维码 ASCII art + 实时状态事件，必须用 SSE
// （text/event-stream）流式推送。GoFrame 的 controller.Bind 模式只支持
// 一次性返回 res 结构体，所以 LoginSSE 走 internal/controller/channels/channels.go
// 里的原生 ghttp.Request handler。LoginReq 在这里只是给前端文档参考用。
package v1
