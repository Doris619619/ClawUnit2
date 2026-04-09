package instances

import (
	"context"

	v1 "clawunit.cuhksz/api/instances/v1"
	"clawunit.cuhksz/internal/middlewares"
	"clawunit.cuhksz/internal/service/lifecycle"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) GetStatus(ctx context.Context, req *v1.GetStatusReq) (res *v1.GetStatusRes, err error) {
	ownerUpn := middlewares.OwnerFromCtx(ctx)

	status, err := lifecycle.GetStatus(ctx, ownerUpn, req.Id)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询实例状态失败")
	}

	return &v1.GetStatusRes{
		Status:   status.Status,
		PodPhase: status.PodPhase,
		Ready:    status.Ready,
	}, nil
}
