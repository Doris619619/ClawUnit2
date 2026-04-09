package transfer

import (
	"context"

	v1 "clawunit.cuhksz/api/transfer/v1"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// Import 通过 K8s exec 执行 tar 导入 OpenClaw 配置
// TODO: 实现配置导入
func (c *ControllerV1) Import(_ context.Context, _ *v1.ImportReq) (res *v1.ImportRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented, "配置导入功能开发中")
}
