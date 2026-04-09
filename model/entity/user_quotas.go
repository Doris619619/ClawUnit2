// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT. Created at 2026-04-07 11:36:13
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// UserQuotas is the golang structure for table user_quotas.
type UserQuotas struct {
	Id           int64       `json:"id"           orm:"id"             description:""` //
	OwnerUpn     string      `json:"ownerUpn"     orm:"owner_upn"      description:""` //
	MaxInstances int32       `json:"maxInstances" orm:"max_instances"  description:""` //
	MaxCpuCores  int32       `json:"maxCpuCores"  orm:"max_cpu_cores"  description:""` //
	MaxMemoryGb  int32       `json:"maxMemoryGb"  orm:"max_memory_gb"  description:""` //
	MaxStorageGb int32       `json:"maxStorageGb" orm:"max_storage_gb" description:""` //
	MaxGpuCount  int32       `json:"maxGpuCount"  orm:"max_gpu_count"  description:""` //
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"     description:""` //
	UpdatedAt    *gtime.Time `json:"updatedAt"    orm:"updated_at"     description:""` //
}
