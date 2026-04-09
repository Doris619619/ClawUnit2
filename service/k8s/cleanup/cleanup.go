package cleanup

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// deleteBySelector 通用的资源删除模板（被 DeleteAllResources 使用）
func deleteBySelector(ctx context.Context, resourceType string, fn func() error) {
	if err := fn(); err != nil {
		g.Log().Warningf(ctx, "删除 %s 时出错: %v", resourceType, err)
	}
}
