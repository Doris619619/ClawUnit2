// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT. Created at 2026-04-07 11:36:13
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// UserQuotas is the golang structure of table user_quotas for DAO operations like Where/Data.
type UserQuotas struct {
	g.Meta       `orm:"table:user_quotas, do:true"`
	Id           any         //
	OwnerUpn     any         //
	MaxInstances any         //
	MaxCpuCores  any         //
	MaxMemoryGb  any         //
	MaxStorageGb any         //
	MaxGpuCount  any         //
	CreatedAt    *gtime.Time //
	UpdatedAt    *gtime.Time //
}
