package skills

import (
	"context"

	v1 "clawunit.cuhksz/api/skills/v1"
	"clawunit.cuhksz/internal/dao"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) ListSystem(ctx context.Context, _ *v1.ListSystemReq) (res *v1.ListSystemRes, err error) {
	var items []*v1.SkillItem

	err = dao.Skills.Ctx(ctx).
		Where("scope", "system").
		Where("enabled", true).
		OrderAsc("name").
		Scan(&items)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询系统技能失败")
	}

	return &v1.ListSystemRes{List: items}, nil
}
