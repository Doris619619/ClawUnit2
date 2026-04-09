package skills

import (
	"context"

	v1 "clawunit.cuhksz/api/skills/v1"
	"clawunit.cuhksz/internal/dao"
	"clawunit.cuhksz/internal/middlewares"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) ListUser(ctx context.Context, _ *v1.ListUserReq) (res *v1.ListUserRes, err error) {
	ownerUpn := middlewares.OwnerFromCtx(ctx)

	var items []*v1.SkillItem

	err = dao.Skills.Ctx(ctx).
		Where("scope", "user").
		Where("owner_upn", ownerUpn).
		OrderAsc("name").
		Scan(&items)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询用户技能失败")
	}

	return &v1.ListUserRes{List: items}, nil
}
