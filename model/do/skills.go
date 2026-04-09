// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT. Created at 2026-04-07 11:36:13
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Skills is the golang structure of table skills for DAO operations like Where/Data.
type Skills struct {
	g.Meta      `orm:"table:skills, do:true"`
	Id          any         //
	Name        any         //
	Description any         //
	Scope       any         //
	OwnerUpn    any         //
	PvcPath     any         //
	Version     any         //
	Enabled     any         //
	CreatedAt   *gtime.Time //
	UpdatedAt   *gtime.Time //
}
