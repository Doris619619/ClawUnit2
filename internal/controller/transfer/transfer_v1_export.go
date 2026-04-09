package transfer

import (
	"context"

	v1 "clawunit.cuhksz/api/transfer/v1"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// Export 通过 K8s exec 执行 tar 导出 OpenClaw 配置
// TODO: 实现配置导出
func (c *ControllerV1) Export(_ context.Context, _ *v1.ExportReq) (res *v1.ExportRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented, "配置导出功能开发中")
}
