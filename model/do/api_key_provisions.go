// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT. Created at 2026-04-07 11:36:13
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// ApiKeyProvisions is the golang structure of table api_key_provisions for DAO operations like Where/Data.
type ApiKeyProvisions struct {
	g.Meta     `orm:"table:api_key_provisions, do:true"`
	Id         any         //
	InstanceId any         //
	OwnerUpn   any         //
	ApiKeyHash any         //
	QuotaPool  any         //
	CreatedAt  *gtime.Time //
	RevokedAt  *gtime.Time //
}
