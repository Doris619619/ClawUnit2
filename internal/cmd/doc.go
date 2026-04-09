// Package cmd 是 ClawUnit HTTP 服务的进程入口。
//
// main.go 在 main 函数里只做一件事：调用 cmd.Main.Run。所有真实的启动
// 工作（路由注册、中间件挂载、后台 goroutine 启动）都集中在这个包，
// 这样 main 包保持空白，不依赖业务代码，方便单测和工具链分析。
//
// # 启动流程
//
// 启动时按下面的顺序发生：
//
//  1. 打印 banner（用 forbidigo 例外允许 fmt.Println，因为这只发生在启动时）。
//  2. 启动 sync.Start 后台 goroutine —— 它会先初始化 K8s client，
//     然后每 5 秒同步一次实例 Pod 状态到数据库。
//  3. 创建 GoFrame Server，把 client max body size 调到 512MB
//     （为 transfer.Import 上传 OpenClaw 配置 tar.gz 留余量）。
//  4. 注册路由组（见下一节）。
//  5. 调用 s.Run() 阻塞主 goroutine。
//
// # 路由组与中间件
//
// 服务总共有 7 个路由组，每个组的中间件链精心设计过：
//
//	路径                             中间件                                  说明
//	-------------------------------- --------------------------------------- --------------------------------------
//	/api/instances/v1                UniRes + InjectIdentity                 实例 CRUD + 生命周期 + 配置热更新
//	/api/skills/v1                   UniRes + InjectIdentity                 技能管理（系统级只读 + 用户级读写）
//	/api/transfer/v1                 UniRes + InjectIdentity                 OpenClaw 配置导入导出
//	/api/channels/v1                 UniRes + InjectIdentity                 渠道插件管理（POST /login 是 SSE，单独注册）
//	/api/admin/v1                    UniRes + InjectIdentity + RequireAdmin  管理员接口（全局实例列表、配额、同步）
//	/api/chat/v1/ws                  无                                      WebSocket 聊天桥（handler 内部认证）
//	/api/gateway/v1/media/...        UniRes + InjectIdentity                 媒体代理（浏览器截图等图片渲染）
//
// 为什么 /api/chat/v1/ws 不挂中间件：浏览器 WebSocket 不会通过 fetch
// 的 Authorization header 发起请求（因为 ws:// upgrade 走的是 HTTP），
// 所以通过 query string ?userId=xxx 传递身份，由 proxy.WsChat 内部
// 自己校验。UniResMiddleware 也会破坏 WebSocket 升级响应，必须绕开。
//
// 为什么 /api/channels/v1/login 走单独的 group.POST 而不是 controller.Bind：
// 它需要返回 SSE 流（text/event-stream），controller 接口模式只支持
// 一次性返回 res 结构体，无法表达流式响应，所以走原生 ghttp.Request handler。
//
// # 修改路由的注意事项
//
//   - 修改 controller 的 request struct（api/<module>/v1/*.go）后必须跑
//     `gf gen ctrl`，它会重新生成 api/<module>/<module>.go 的 interface
//     和 internal/controller/<module>/<module>_v1_<method>.go 文件。
//   - 不要把 API 处理函数塞到 internal/controller/<module>/<module>.go
//     里 —— 那是给包级 helper 用的（例如 channels.LoginSSE）。
//   - 新增路由组时复制现有写法即可，注意：除非你明确知道为什么不要，
//     否则永远先挂 UniResMiddleware，再挂 InjectIdentity。顺序不能反，
//     否则 InjectIdentity 抛出的错误不会被 UniResMiddleware 捕获包装成
//     统一响应格式。
package cmd
