package skills

import (
	"context"

	v1 "clawunit.cuhksz/api/skills/v1"
	"clawunit.cuhksz/internal/dao"
	"clawunit.cuhksz/internal/middlewares"
	"clawunit.cuhksz/internal/model/do"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) Upload(ctx context.Context, req *v1.UploadReq) (res *v1.UploadRes, err error) {
	ownerUpn := middlewares.OwnerFromCtx(ctx)
	pvcPath := ".skills/" + req.Name

	id, err := dao.Skills.Ctx(ctx).Data(do.Skills{
		Name:        req.Name,
		Description: req.Description,
		Scope:       "user",
		OwnerUpn:    ownerUpn,
		PvcPath:     pvcPath,
		Enabled:     true,
	}).InsertAndGetId()
	if err != nil {
		return nil, gerror.Wrapf(err, "保存技能记录失败")
	}

	// TODO: 通过 K8s exec 将文件写入用户 PVC

	return &v1.UploadRes{Id: id}, nil
}
