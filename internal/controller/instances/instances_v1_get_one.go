package instances

import (
	"context"

	v1 "clawunit.cuhksz/api/instances/v1"
	"clawunit.cuhksz/internal/dao"
	"clawunit.cuhksz/internal/middlewares"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) GetOne(ctx context.Context, req *v1.GetOneReq) (res *v1.GetOneRes, err error) {
	ownerUpn := middlewares.OwnerFromCtx(ctx)

	var item v1.InstanceItem

	err = dao.Instances.Ctx(ctx).
		Where("id", req.Id).
		Where("owner_upn", ownerUpn).
		Scan(&item)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询实例失败")
	}

	if item.Id == 0 {
		return nil, gerror.NewCodef(gcode.CodeNotFound, "实例 %d 不存在", req.Id)
	}

	return &v1.GetOneRes{Item: &item}, nil
}
