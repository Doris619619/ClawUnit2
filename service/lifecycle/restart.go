package lifecycle

import (
	"context"
	"fmt"
)

// Restart 等价于 Stop 然后 Start。
//
// 用于强制刷新 Pod 进程状态（例如配置变更后让 init container 重新
// 复制 ConfigMap 到 PVC）。注意 Stop 不会等 Pod 完全终止，所以 Start
// 创建 Pod 时可能会撞 AlreadyExists —— pod.Create 内部已经处理这种
// 情况（删除旧 Pod 重建）。
func Restart(ctx context.Context, ownerUpn string, instanceID int64) error {
	if err := Stop(ctx, ownerUpn, instanceID); err != nil {
		return fmt.Errorf("停止实例失败: %w", err)
	}

	return Start(ctx, ownerUpn, instanceID)
}
