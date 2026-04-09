package instances

import (
	"context"

	v1 "clawunit.cuhksz/api/instances/v1"
	"clawunit.cuhksz/internal/middlewares"
	"clawunit.cuhksz/internal/service/lifecycle"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) Restart(ctx context.Context, req *v1.RestartReq) (res *v1.RestartRes, err error) {
	ownerUpn := middlewares.OwnerFromCtx(ctx)

	if err = lifecycle.Restart(ctx, ownerUpn, req.Id); err != nil {
		return nil, gerror.Wrapf(err, "重启实例失败")
	}

	return &v1.RestartRes{}, nil
}
