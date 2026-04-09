package instances

import (
	"context"

	v1 "clawunit.cuhksz/api/instances/v1"
	"clawunit.cuhksz/internal/middlewares"
	"clawunit.cuhksz/internal/service/lifecycle"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) Stop(ctx context.Context, req *v1.StopReq) (res *v1.StopRes, err error) {
	ownerUpn := middlewares.OwnerFromCtx(ctx)

	if err = lifecycle.Stop(ctx, ownerUpn, req.Id); err != nil {
		return nil, gerror.Wrapf(err, "停止实例失败")
	}

	return &v1.StopRes{}, nil
}
