// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package skills

import (
	"context"

	"clawunit.cuhksz/api/skills/v1"
)

type ISkillsV1 interface {
	ListSystem(ctx context.Context, req *v1.ListSystemReq) (res *v1.ListSystemRes, err error)
	ListUser(ctx context.Context, req *v1.ListUserReq) (res *v1.ListUserRes, err error)
	Upload(ctx context.Context, req *v1.UploadReq) (res *v1.UploadRes, err error)
	DeleteSkill(ctx context.Context, req *v1.DeleteSkillReq) (res *v1.DeleteSkillRes, err error)
}
