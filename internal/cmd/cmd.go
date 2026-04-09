package cmd

import (
	"context"
	_ "embed"
	"fmt"

	"clawunit.cuhksz/internal/controller/admin"
	"clawunit.cuhksz/internal/controller/channels"
	"clawunit.cuhksz/internal/controller/instances"
	"clawunit.cuhksz/internal/controller/skills"
	"clawunit.cuhksz/internal/controller/transfer"
	"clawunit.cuhksz/internal/middlewares"
	"clawunit.cuhksz/internal/service/proxy"
	"clawunit.cuhksz/internal/service/sync"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
)

// Main 是 ClawUnit 的根 CLI 命令，由 main.main 调用。
// 它启动 K8s 同步后台 goroutine 然后启动 HTTP 服务器。
var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start clawunit http server",
		Func: func(ctx context.Context, _ *gcmd.Parser) (err error) {
			//nolint:forbidigo
			{
				fmt.Println("  ____ _                _   _       _ _   ")
				fmt.Println(" / ___| | __ ___      _| | | |_ __ (_) |_ ")
				fmt.Println("| |   | |/ _` \\ \\ /\\ / / | | | '_ \\| | __|")
				fmt.Println("| |___| | (_| |\\ V  /| |_| |")
				fmt.Println(" \\____|_|\\__,_| \\_/\\_/  \\___/|_| |_|_|\\__|")
				fmt.Println("ClawUnit - OpenClaw Instance Manager")
				fmt.Println("Copyright 2026 The Chinese University of Hong Kong, Shenzhen")
				fmt.Println()
			}

			// 启动 K8s 状态同步服务
			go sync.Start(ctx)

			s := g.Server()
			s.SetClientMaxBodySize(512 * 1024 * 1024) // 512 MB
			s.SetPort(g.Cfg().MustGet(ctx, "server.httpPort").Int())

			// 实例管理（需要用户身份）
			s.Group("/api/instances/v1", func(group *ghttp.RouterGroup) {
				group.Middleware(middlewares.UniResMiddleware, middlewares.InjectIdentity)
				group.Bind(instances.NewV1())
			})

			// 技能管理
			s.Group("/api/skills/v1", func(group *ghttp.RouterGroup) {
				group.Middleware(middlewares.UniResMiddleware, middlewares.InjectIdentity)
				group.Bind(skills.NewV1())
			})

			// 配置导入导出
			s.Group("/api/transfer/v1", func(group *ghttp.RouterGroup) {
				group.Middleware(middlewares.UniResMiddleware, middlewares.InjectIdentity)
				group.Bind(transfer.NewV1())
			})

			// 渠道管理（插件安装、扫码登录等）
			s.Group("/api/channels/v1", func(group *ghttp.RouterGroup) {
				group.Middleware(middlewares.UniResMiddleware, middlewares.InjectIdentity)
				group.Bind(channels.NewV1())
				group.POST("/login", channels.LoginSSE)
			})

			// 管理员接口
			s.Group("/api/admin/v1", func(group *ghttp.RouterGroup) {
				group.Middleware(middlewares.UniResMiddleware, middlewares.InjectIdentity, middlewares.RequireAdmin)
				group.Bind(admin.NewV1())
			})

			// WebSocket 聊天（不走中间件，handler 内部认证）
			s.Group("/api/chat/v1", func(group *ghttp.RouterGroup) {
				group.GET("/ws", proxy.WsChat)
			})

			// 媒体代理（截图/生成图片等，需要用户身份）
			s.Group("/api/gateway/v1", func(group *ghttp.RouterGroup) {
				group.Middleware(middlewares.UniResMiddleware, middlewares.InjectIdentity)
				group.GET("/media/{instanceId}/{mediaType}", proxy.LatestMedia)
			})

			s.Run()

			return nil
		},
	}
)
