package pod

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// DeleteAndWait 删除 Pod 并等待完全终止（最多 30 秒）
func DeleteAndWait(ctx context.Context, ownerUpn string, instanceID int64) error {
	if err := Delete(ctx, ownerUpn, instanceID); err != nil {
		return err
	}

	// 轮询等待 Pod 消失
	for range 15 {
		time.Sleep(2 * time.Second)

		_, err := Get(ctx, ownerUpn, instanceID)
		if err != nil {
			// Pod 不存在了
			return nil
		}
	}

	g.Log().Warningf(ctx, "实例 %d 的 Pod 终止超时（30s），继续清理", instanceID)

	return nil
}
