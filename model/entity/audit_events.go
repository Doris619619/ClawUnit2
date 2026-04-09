// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT. Created at 2026-04-07 11:36:13
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/os/gtime"
)

// AuditEvents is the golang structure for table audit_events.
type AuditEvents struct {
	Id           int64       `json:"id"           orm:"id"            description:""` //
	ActorUpn     string      `json:"actorUpn"     orm:"actor_upn"     description:""` //
	Action       string      `json:"action"       orm:"action"        description:""` //
	ResourceType string      `json:"resourceType" orm:"resource_type" description:""` //
	ResourceId   int64       `json:"resourceId"   orm:"resource_id"   description:""` //
	Details      *gjson.Json `json:"details"      orm:"details"       description:""` //
	IpAddress    string      `json:"ipAddress"    orm:"ip_address"    description:""` //
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"    description:""` //
}
