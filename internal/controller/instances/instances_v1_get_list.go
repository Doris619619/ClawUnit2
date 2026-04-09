package instances

import (
	"context"

	v1 "clawunit.cuhksz/api/instances/v1"
	"clawunit.cuhksz/internal/dao"
	"clawunit.cuhksz/internal/middlewares"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) GetList(ctx context.Context, req *v1.GetListReq) (res *v1.GetListRes, err error) {
	ownerUpn := middlewares.OwnerFromCtx(ctx)

	model := dao.Instances.Ctx(ctx).Where("owner_upn", ownerUpn)
	if req.Status != "" {
		model = model.Where("status", req.Status)
	}

	total, err := model.Clone().Count()
	if err != nil {
		return nil, gerror.Wrapf(err, "查询实例总数失败")
	}

	var items []*v1.InstanceItem

	err = model.
		Page(req.Page, req.PageSize).
		OrderDesc("created_at").
		Scan(&items)
	if err != nil {
		if items == nil {
			items = make([]*v1.InstanceItem, 0)
		}
	}

	return &v1.GetListRes{List: items, Total: total}, nil
}
