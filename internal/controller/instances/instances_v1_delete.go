package instances

import (
	"context"

	v1 "clawunit.cuhksz/api/instances/v1"
	"clawunit.cuhksz/internal/dao"
	"clawunit.cuhksz/internal/middlewares"
	"clawunit.cuhksz/internal/service/lifecycle"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	ownerUpn := middlewares.OwnerFromCtx(ctx)

	count, err := dao.Instances.Ctx(ctx).Where("id", req.Id).Where("owner_upn", ownerUpn).Count()
	if err != nil {
		return nil, gerror.Wrapf(err, "查询实例失败")
	}

	if count == 0 {
		return nil, gerror.NewCodef(gcode.CodeNotFound, "实例 %d 不存在", req.Id)
	}

	if err = lifecycle.Delete(ctx, ownerUpn, req.Id); err != nil {
		return nil, gerror.Wrapf(err, "删除实例失败")
	}

	return &v1.DeleteRes{}, nil
}
