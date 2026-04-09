// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT. Created at 2026-04-07 11:36:13
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ApiKeyProvisions is the golang structure for table api_key_provisions.
type ApiKeyProvisions struct {
	Id         int64       `json:"id"         orm:"id"           description:""` //
	InstanceId int64       `json:"instanceId" orm:"instance_id"  description:""` //
	OwnerUpn   string      `json:"ownerUpn"   orm:"owner_upn"    description:""` //
	ApiKeyHash string      `json:"apiKeyHash" orm:"api_key_hash" description:""` //
	QuotaPool  string      `json:"quotaPool"  orm:"quota_pool"   description:""` //
	CreatedAt  *gtime.Time `json:"createdAt"  orm:"created_at"   description:""` //
	RevokedAt  *gtime.Time `json:"revokedAt"  orm:"revoked_at"   description:""` //
}
