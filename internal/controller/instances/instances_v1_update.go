package instances

import (
	"context"

	v1 "clawunit.cuhksz/api/instances/v1"
	"clawunit.cuhksz/internal/dao"
	"clawunit.cuhksz/internal/middlewares"
	"clawunit.cuhksz/internal/model/do"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	ownerUpn := middlewares.OwnerFromCtx(ctx)

	count, err := dao.Instances.Ctx(ctx).Where("id", req.Id).Where("owner_upn", ownerUpn).Count()
	if err != nil {
		return nil, gerror.Wrapf(err, "查询实例失败")
	}

	if count == 0 {
		return nil, gerror.NewCodef(gcode.CodeNotFound, "实例 %d 不存在", req.Id)
	}

	updateDO := do.Instances{}
	if req.Name != nil {
		updateDO.Name = *req.Name
	}

	if req.Description != nil {
		updateDO.Description = *req.Description
	}

	if updateDO.Name == nil && updateDO.Description == nil {
		return &v1.UpdateRes{Affected: 0}, nil
	}

	result, err := dao.Instances.Ctx(ctx).Where("id", req.Id).Data(updateDO).Update()
	if err != nil {
		return nil, gerror.Wrapf(err, "更新实例失败")
	}

	affected, _ := result.RowsAffected()

	return &v1.UpdateRes{Affected: affected}, nil
}
