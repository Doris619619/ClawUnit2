package skills

import (
	"context"

	v1 "clawunit.cuhksz/api/skills/v1"
	"clawunit.cuhksz/internal/dao"
	"clawunit.cuhksz/internal/middlewares"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) DeleteSkill(ctx context.Context, req *v1.DeleteSkillReq) (res *v1.DeleteSkillRes, err error) {
	ownerUpn := middlewares.OwnerFromCtx(ctx)

	result, err := dao.Skills.Ctx(ctx).
		Where("id", req.Id).
		Where("owner_upn", ownerUpn).
		Where("scope", "user").
		Delete()
	if err != nil {
		return nil, gerror.Wrapf(err, "删除技能失败")
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return nil, gerror.NewCodef(gcode.CodeNotFound, "技能 %d 不存在", req.Id)
	}

	// TODO: 通过 K8s exec 从用户 PVC 中删除文件

	return &v1.DeleteSkillRes{}, nil
}
