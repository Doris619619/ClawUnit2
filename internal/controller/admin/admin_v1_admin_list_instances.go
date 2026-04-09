package admin

import (
	"context"

	v1 "clawunit.cuhksz/api/admin/v1"
	"clawunit.cuhksz/internal/dao"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) AdminListInstances(ctx context.Context, req *v1.AdminListInstancesReq) (res *v1.AdminListInstancesRes, err error) {
	model := dao.Instances.Ctx(ctx)

	if req.OwnerUpn != "" {
		model = model.Where("owner_upn", req.OwnerUpn)
	}

	if req.Status != "" {
		model = model.Where("status", req.Status)
	}

	total, err := model.Clone().Count()
	if err != nil {
		return nil, gerror.Wrapf(err, "查询实例总数失败")
	}

	var items []*v1.AdminInstanceItem

	err = model.
		Page(req.Page, req.PageSize).
		OrderDesc("created_at").
		Scan(&items)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询实例列表失败")
	}

	return &v1.AdminListInstancesRes{List: items, Total: total}, nil
}
