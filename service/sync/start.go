package sync

import (
	"context"
	"time"

	"clawunit.cuhksz/internal/service/k8s"

	"github.com/gogf/gf/v2/frame/g"
)

// Start 启动 K8s 状态同步后台 goroutine。
//
// 调用方应该用 go sync.Start(ctx) 在应用启动时调用一次。
// 函数会先初始化 K8s client，失败时记 error 日志后直接返回（不会
// panic，也不会重试） —— 这种情况下 ClawUnit HTTP 服务还能起来，
// 但所有依赖 K8s 的接口都会失败。
//
// 同步周期是 syncInterval 常量定义的 5 秒，通过 ctx 取消停止。
func Start(ctx context.Context) {
	// 初始化 K8s 客户端
	if err := k8s.Initialize(ctx); err != nil {
		g.Log().Errorf(ctx, "K8s 客户端初始化失败，同步服务未启动: %v", err)

		return
	}

	g.Log().Info(ctx, "K8s 状态同步服务已启动")

	ticker := time.NewTicker(syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			g.Log().Info(ctx, "K8s 状态同步服务已停止")

			return
		case <-ticker.C:
			syncInstanceStatus(ctx)
		}
	}
}
