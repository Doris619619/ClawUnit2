// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT. Created at 2026-04-07 11:36:13
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Skills is the golang structure for table skills.
type Skills struct {
	Id          int64       `json:"id"          orm:"id"          description:""` //
	Name        string      `json:"name"        orm:"name"        description:""` //
	Description string      `json:"description" orm:"description" description:""` //
	Scope       string      `json:"scope"       orm:"scope"       description:""` //
	OwnerUpn    string      `json:"ownerUpn"    orm:"owner_upn"   description:""` //
	PvcPath     string      `json:"pvcPath"     orm:"pvc_path"    description:""` //
	Version     string      `json:"version"     orm:"version"     description:""` //
	Enabled     bool        `json:"enabled"     orm:"enabled"     description:""` //
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"  description:""` //
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"  description:""` //
}
