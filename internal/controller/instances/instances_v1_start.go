package instances

import (
	"context"

	v1 "clawunit.cuhksz/api/instances/v1"
	"clawunit.cuhksz/internal/middlewares"
	"clawunit.cuhksz/internal/service/lifecycle"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) Start(ctx context.Context, req *v1.StartReq) (res *v1.StartRes, err error) {
	ownerUpn := middlewares.OwnerFromCtx(ctx)

	if err = lifecycle.Start(ctx, ownerUpn, req.Id); err != nil {
		return nil, gerror.Wrapf(err, "启动实例失败")
	}

	return &v1.StartRes{}, nil
}
