package instances

import (
	"context"

	v1 "clawunit.cuhksz/api/instances/v1"
	"clawunit.cuhksz/internal/dao"
	"clawunit.cuhksz/internal/middlewares"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) GetQuota(ctx context.Context, _ *v1.GetQuotaReq) (res *v1.GetQuotaRes, err error) {
	ownerUpn := middlewares.OwnerFromCtx(ctx)

	quota, err := dao.UserQuotas.Ctx(ctx).Where("owner_upn", ownerUpn).One()
	if err != nil {
		return nil, gerror.Wrapf(err, "查询配额失败")
	}

	item := &v1.QuotaItem{OwnerUpn: ownerUpn}

	if !quota.IsEmpty() {
		item.MaxInstances = quota["max_instances"].Int32()
	} else {
		item.MaxInstances = 3
	}

	// 前期资源固定，只统计实例数量
	usedCount, err := dao.Instances.Ctx(ctx).
		Where("owner_upn", ownerUpn).
		WhereNotIn("status", []string{"deleting", "error"}).
		Count()
	if err != nil {
		return nil, gerror.Wrapf(err, "查询使用量失败")
	}

	item.UsedInstances = usedCount

	return &v1.GetQuotaRes{Quota: item}, nil
}
