package sync

import "context"

// ForceSync 立即同步一次，不等待下一个 ticker。
//
// 给 admin API 用 —— 通常是修改了实例数据后想立刻看到 status
// 推进，不想等 5 秒。同步是同步执行的，调用会阻塞直到完成。
func ForceSync(ctx context.Context) {
	syncInstanceStatus(ctx)
}
