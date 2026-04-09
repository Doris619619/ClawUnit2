// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT. Created at 2026-04-07 11:36:13
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Instances is the golang structure for table instances.
type Instances struct {
	Id                  int64       `json:"id"                  orm:"id"                    description:""` //
	OwnerUpn            string      `json:"ownerUpn"            orm:"owner_upn"             description:""` //
	Name                string      `json:"name"                orm:"name"                  description:""` //
	Description         string      `json:"description"         orm:"description"           description:""` //
	Status              string      `json:"status"              orm:"status"                description:""` //
	Image               string      `json:"image"               orm:"image"                 description:""` //
	StorageClass        string      `json:"storageClass"        orm:"storage_class"         description:""` //
	MountPath           string      `json:"mountPath"           orm:"mount_path"            description:""` //
	PodName             string      `json:"podName"             orm:"pod_name"              description:""` //
	PodNamespace        string      `json:"podNamespace"        orm:"pod_namespace"         description:""` //
	PodIp               string      `json:"podIp"               orm:"pod_ip"                description:""` //
	AccessToken         string      `json:"accessToken"         orm:"access_token"          description:""` //
	ApiKeyHash          string      `json:"apiKeyHash"          orm:"api_key_hash"          description:""` //
	CpuCores            string      `json:"cpuCores"            orm:"cpu_cores"             description:""` //
	MemoryGb            string      `json:"memoryGb"            orm:"memory_gb"             description:""` //
	DiskGb              string      `json:"diskGb"              orm:"disk_gb"               description:""` //
	GpuCount            int32       `json:"gpuCount"            orm:"gpu_count"             description:""` //
	ContainerPort       int32       `json:"containerPort"       orm:"container_port"        description:""` //
	GpuEnabled          bool        `json:"gpuEnabled"          orm:"gpu_enabled"           description:""` //
	CreatedAt           *gtime.Time `json:"createdAt"           orm:"created_at"            description:""` //
	UpdatedAt           *gtime.Time `json:"updatedAt"           orm:"updated_at"            description:""` //
	StartedAt           *gtime.Time `json:"startedAt"           orm:"started_at"            description:""` //
	StoppedAt           *gtime.Time `json:"stoppedAt"           orm:"stopped_at"            description:""` //
	ApiMode             string      `json:"apiMode"             orm:"api_mode"              description:""` //
	ApiKey              string      `json:"apiKey"              orm:"api_key"               description:""` //
	BaseUrl             string      `json:"baseUrl"             orm:"base_url"              description:""` //
	GatewayToken        string      `json:"gatewayToken"        orm:"gateway_token"         description:""` //
	Provider            string      `json:"provider"            orm:"provider"              description:""` //
	ModelId             string      `json:"modelId"             orm:"model_id"              description:""` //
	AllowPrivateNetwork bool        `json:"allowPrivateNetwork" orm:"allow_private_network" description:""` //
	PvcName             string      `json:"pvcName"             orm:"pvc_name"              description:""` //
}
