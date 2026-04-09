package admin

import (
	"context"

	v1 "clawunit.cuhksz/api/admin/v1"
	"clawunit.cuhksz/internal/dao"
	"clawunit.cuhksz/internal/model/do"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) AdminUpdateQuota(ctx context.Context, req *v1.AdminUpdateQuotaReq) (res *v1.AdminUpdateQuotaRes, err error) {
	quota := do.UserQuotas{OwnerUpn: req.OwnerUpn}

	if req.MaxInstances != nil {
		quota.MaxInstances = *req.MaxInstances
	}

	count, err := dao.UserQuotas.Ctx(ctx).Where("owner_upn", req.OwnerUpn).Count()
	if err != nil {
		return nil, gerror.Wrapf(err, "查询配额失败")
	}

	if count > 0 {
		_, err = dao.UserQuotas.Ctx(ctx).Where("owner_upn", req.OwnerUpn).Data(quota).Update()
	} else {
		_, err = dao.UserQuotas.Ctx(ctx).Data(quota).Insert()
	}

	if err != nil {
		return nil, gerror.Wrapf(err, "更新配额失败")
	}

	return &v1.AdminUpdateQuotaRes{}, nil
}
