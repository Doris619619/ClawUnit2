package instances

import (
	"context"

	v1 "clawunit.cuhksz/api/instances/v1"
	"clawunit.cuhksz/internal/dao"
	"clawunit.cuhksz/internal/middlewares"
	"clawunit.cuhksz/internal/service/k8s/pvc"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) DeletePVC(ctx context.Context, req *v1.DeletePVCReq) (res *v1.DeletePVCRes, err error) {
	ownerUpn := middlewares.OwnerFromCtx(ctx)

	// 确认 PVC 不属于任何活跃实例
	count, err := dao.Instances.Ctx(ctx).Where("owner_upn", ownerUpn).Where("pvc_name", req.Name).Count()
	if err != nil {
		return nil, gerror.Wrapf(err, "查询实例失败")
	}

	if count > 0 {
		return nil, gerror.NewCodef(gcode.CodeNotAuthorized, "该数据正在被实例使用，无法删除")
	}

	if err = pvc.DeleteInstance(ctx, ownerUpn, req.Name); err != nil {
		return nil, gerror.Wrapf(err, "删除数据失败")
	}

	return &v1.DeletePVCRes{}, nil
}
