// Command clawunit 启动 ClawUnit HTTP 服务。
//
// ClawUnit 是 CUHKSZ AI 平台的 OpenClaw 实例管理服务，负责在 Kubernetes
// 上创建/调度/管理 OpenClaw AI 助手实例，并代理前端到 OpenClaw 的
// HTTP 与 WebSocket 流量。
//
// 入口逻辑放在 internal/cmd 包，本文件保持空白以便 main 包不直接
// 依赖业务代码，方便单测。运行：
//
//	go run main.go
//
// 配置文件路径：manifest/config/config.yaml。
package main

import (
	"github.com/gogf/gf/v2/os/gctx"

	"clawunit.cuhksz/internal/cmd"

	// 注册 PostgreSQL 驱动到 GoFrame ORM
	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
