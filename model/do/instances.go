// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT. Created at 2026-04-07 11:36:13
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Instances is the golang structure of table instances for DAO operations like Where/Data.
type Instances struct {
	g.Meta              `orm:"table:instances, do:true"`
	Id                  any         //
	OwnerUpn            any         //
	Name                any         //
	Description         any         //
	Status              any         //
	Image               any         //
	StorageClass        any         //
	MountPath           any         //
	PodName             any         //
	PodNamespace        any         //
	PodIp               any         //
	AccessToken         any         //
	ApiKeyHash          any         //
	CpuCores            any         //
	MemoryGb            any         //
	DiskGb              any         //
	GpuCount            any         //
	ContainerPort       any         //
	GpuEnabled          any         //
	CreatedAt           *gtime.Time //
	UpdatedAt           *gtime.Time //
	StartedAt           *gtime.Time //
	StoppedAt           *gtime.Time //
	ApiMode             any         //
	ApiKey              any         //
	BaseUrl             any         //
	GatewayToken        any         //
	Provider            any         //
	ModelId             any         //
	AllowPrivateNetwork any         //
	PvcName             any         //
}
