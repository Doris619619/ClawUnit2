package instances

import (
	"context"

	v1 "clawunit.cuhksz/api/instances/v1"
	"clawunit.cuhksz/internal/middlewares"
	"clawunit.cuhksz/internal/service/lifecycle"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	ownerUpn := middlewares.OwnerFromCtx(ctx)

	result, err := lifecycle.Create(ctx, lifecycle.CreateRequest{
		Name:                req.Name,
		Description:         req.Description,
		OwnerUpn:            ownerUpn,
		Image:               req.Image,
		StorageClass:        req.StorageClass,
		GPUCount:            req.GPUCount,
		GPUEnabled:          req.GPUCount > 0,
		ApiMode:             req.ApiMode,
		ModelID:             req.ModelID,
		ApiKey:              req.ApiKey,
		BaseUrl:             req.BaseUrl,
		QuotaPool:           req.QuotaPool,
		AllowPrivateNetwork: req.AllowPrivateNetwork,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "创建实例失败")
	}

	return &v1.CreateRes{Id: result.InstanceID}, nil
}
