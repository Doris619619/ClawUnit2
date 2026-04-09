package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SkillItem 技能信息
type SkillItem struct {
	CreatedAt   *gtime.Time `json:"createdAt" dc:"创建时间"`
	Name        string      `json:"name" dc:"技能名称"`
	Description string      `json:"description" dc:"技能描述"`
	Scope       string      `json:"scope" dc:"作用域：system/user"`
	PvcPath     string      `json:"pvcPath" dc:"PVC 中的路径"`
	Version     string      `json:"version" dc:"版本号"`
	Id          int64       `json:"id" dc:"技能ID"`
	Enabled     bool        `json:"enabled" dc:"是否启用"`
}

// ListSystemReq 列出系统技能
type ListSystemReq struct {
	g.Meta `path:"/system/list" method:"get" tags:"Skills" summary:"获取系统技能列表" dc:"列出管理员预置的系统技能。"`
}

type ListSystemRes struct {
	List []*SkillItem `json:"list" dc:"系统技能列表"`
}

// ListUserReq 列出用户自定义技能
type ListUserReq struct {
	g.Meta `path:"/user/list" method:"get" tags:"Skills" summary:"获取用户技能列表" dc:"列出当前用户的自定义技能。"`
}

type ListUserRes struct {
	List []*SkillItem `json:"list" dc:"用户技能列表"`
}

// UploadReq 上传用户技能
type UploadReq struct {
	g.Meta `path:"/user/upload" method:"post" tags:"Skills" summary:"上传用户技能" dc:"上传技能文件到用户 PVC。"`

	Name        string `json:"name" v:"required|length:1,128" dc:"技能名称"`
	Description string `json:"description" v:"max-length:500" dc:"技能描述"`
	// 文件通过 multipart/form-data 上传
}

type UploadRes struct {
	Id int64 `json:"id" dc:"技能ID"`
}

// DeleteSkillReq 删除用户技能
type DeleteSkillReq struct {
	g.Meta `path:"/user/delete" method:"post" tags:"Skills" summary:"删除用户技能" dc:"从用户 PVC 中删除技能。"`

	Id int64 `json:"id" v:"required|min:1" dc:"技能ID"`
}

type DeleteSkillRes struct{}
