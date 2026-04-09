// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT. Created at 2026-04-07 11:36:13
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// AuditEvents is the golang structure of table audit_events for DAO operations like Where/Data.
type AuditEvents struct {
	g.Meta       `orm:"table:audit_events, do:true"`
	Id           any         //
	ActorUpn     any         //
	Action       any         //
	ResourceType any         //
	ResourceId   any         //
	Details      *gjson.Json //
	IpAddress    any         //
	CreatedAt    *gtime.Time //
}
