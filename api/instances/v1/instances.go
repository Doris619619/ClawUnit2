package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// InstanceItem 是给前端展示的实例信息。
//
// 字段是 instances 表的子集，剔除了不该暴露的内部字段（API key 原文、
// access_token 等）。Status 取值见 GetListReq.Status 的 v 标签。
type InstanceItem struct {
	CreatedAt     *gtime.Time `json:"createdAt" dc:"创建时间"`
	UpdatedAt     *gtime.Time `json:"updatedAt" dc:"更新时间"`
	StartedAt     *gtime.Time `json:"startedAt" dc:"最近启动时间"`
	StoppedAt     *gtime.Time `json:"stoppedAt" dc:"最近停止时间"`
	Name          string      `json:"name" dc:"实例名称"`
	Description   string      `json:"description" dc:"实例描述"`
	Status        string      `json:"status" dc:"实例状态：creating/running/stopped/error/deleting"`
	Image         string      `json:"image" dc:"容器镜像"`
	StorageClass  string      `json:"storageClass" dc:"存储类"`
	MountPath     string      `json:"mountPath" dc:"挂载路径"`
	PodName       string      `json:"podName" dc:"Pod 名称"`
	PodNamespace  string      `json:"podNamespace" dc:"Pod 命名空间"`
	OwnerUpn      string      `json:"ownerUpn" dc:"所属用户 UPN"`
	CpuCores      string      `json:"cpuCores" dc:"CPU 资源量，如 200m"`
	MemoryGb      string      `json:"memoryGb" dc:"内存资源量，如 512Mi"`
	DiskGb        string      `json:"diskGb" dc:"磁盘资源量，如 500Mi"`
	Id            int64       `json:"id" dc:"实例ID"`
	GPUCount      int32       `json:"gpuCount" dc:"GPU 数量"`
	ContainerPort int32       `json:"containerPort" dc:"容器端口"`
	GPUEnabled    bool        `json:"gpuEnabled" dc:"是否启用 GPU"`
}

// QuotaItem 是前端展示的用户配额。
//
// 当前只跟踪实例数量上限；CPU/内存/磁盘等资源量目前是服务端固定配置，
// 不让用户调整，所以也不在 quota 里展示。后期开放资源自定义后再扩展。
type QuotaItem struct {
	OwnerUpn      string `json:"ownerUpn" dc:"用户 UPN"`
	MaxInstances  int32  `json:"maxInstances" dc:"最大实例数"`
	UsedInstances int    `json:"usedInstances" dc:"已用实例数"`
}

// GetListReq 分页查询用户实例
type GetListReq struct {
	g.Meta `path:"/list" method:"get" tags:"Instances" summary:"获取实例列表" dc:"查询当前用户的实例列表。"`

	Status   string `json:"status" v:"in:creating,running,stopped,error,deleting" dc:"按状态过滤"`
	Page     int    `json:"page" d:"1" v:"min:1" dc:"页码"`
	PageSize int    `json:"pageSize" d:"20" v:"min:1|max:100" dc:"每页数量"`
}

type GetListRes struct {
	List  []*InstanceItem `json:"list" dc:"实例列表"`
	Total int             `json:"total" dc:"总数"`
}

// GetOneReq 查询单个实例
type GetOneReq struct {
	g.Meta `path:"/detail" method:"get" tags:"Instances" summary:"获取实例详情" dc:"根据实例 ID 查询详情。"`

	Id int64 `json:"id" v:"required|min:1" dc:"实例ID"`
}

type GetOneRes struct {
	Item *InstanceItem `json:"item" dc:"实例详情"`
}

// CreateReq 创建实例（资源配额使用服务端配置，前期不开放用户自定义）
type CreateReq struct {
	g.Meta `path:"/create" method:"post" tags:"Instances" summary:"创建实例" dc:"创建 OpenClaw 实例并分配 K8s 资源。"`

	ApiMode             string `json:"apiMode" d:"manual" v:"in:auto,manual" dc:"API 配置模式：auto 通过 Open Platform 自动分配，manual 手动填写"`
	Description         string `json:"description" v:"max-length:500" dc:"实例描述"`
	Image               string `json:"image" dc:"容器镜像，不传则使用默认镜像"`
	StorageClass        string `json:"storageClass" dc:"存储类，不传则使用默认"`
	Name                string `json:"name" v:"required|length:3,50" dc:"实例名称，3-50 字符"`
	ModelID             string `json:"modelId" dc:"默认模型 ID，如 qwen/qwen3.6-plus-preview:free"`
	ApiKey              string `json:"apiKey" dc:"手动模式下的 API Key"`
	BaseUrl             string `json:"baseUrl" dc:"手动模式下的 API Base URL"`
	QuotaPool           string `json:"quotaPool" dc:"auto 模式下的配额池名称（前端传入）"`
	ExistingPVC         string `json:"existingPvc" dc:"绑定已有的 PVC 名称（恢复历史数据），不传则创建新 PVC"`
	GPUCount            int32  `json:"gpuCount" d:"0" v:"between:0,4" dc:"GPU 数量，0-4"`
	AllowPrivateNetwork bool   `json:"allowPrivateNetwork" dc:"允许访问私有网络（需管理员授权）"`
}

type CreateRes struct {
	Id int64 `json:"id" dc:"新建实例 ID"`
}

// UpdateReq PATCH 更新实例元数据
type UpdateReq struct {
	g.Meta `path:"/update" method:"post" tags:"Instances" summary:"更新实例" dc:"更新实例名称和描述。"`

	Name        *string `json:"name" v:"length:3,50" dc:"实例名称"`
	Description *string `json:"description" v:"max-length:500" dc:"实例描述"`
	Id          int64   `json:"id" v:"required|min:1" dc:"实例ID"`
}

type UpdateRes struct {
	Affected int64 `json:"affected" dc:"受影响行数"`
}

// UpdateConfigReq 更新实例的 OpenClaw 配置（热生效，无需重启）
type UpdateConfigReq struct {
	g.Meta `path:"/config" method:"post" tags:"Instances" summary:"更新 OpenClaw 配置" dc:"更新搜索引擎、工具权限等配置，实时生效。"`

	ModelID             *string `json:"modelId" dc:"默认模型 ID"`
	ApiKey              *string `json:"apiKey" dc:"LLM API Key"`
	BaseUrl             *string `json:"baseUrl" dc:"LLM Base URL"`
	SearchProvider      *string `json:"searchProvider" dc:"搜索引擎：brave/perplexity/duckduckgo/gemini/kimi 等"`
	SearchApiKey        *string `json:"searchApiKey" dc:"搜索引擎 API Key"`
	ToolProfile         *string `json:"toolProfile" v:"in:full,coding,messaging,minimal" dc:"工具权限级别"`
	SystemPrompt        *string `json:"systemPrompt" dc:"自定义系统提示词（写入 AGENTS.md）"`
	AllowPrivateNetwork *bool   `json:"allowPrivateNetwork" dc:"允许访问私有网络（需管理员授权）"`
	Id                  int64   `json:"id" v:"required|min:1" dc:"实例ID"`
}

type UpdateConfigRes struct{}

// DeleteReq 删除实例
type DeleteReq struct {
	g.Meta `path:"/delete" method:"post" tags:"Instances" summary:"删除实例" dc:"删除实例及其所有 K8s 资源。"`

	Id         int64 `json:"id" v:"required|min:1" dc:"实例ID"`
	DeleteData bool  `json:"deleteData" dc:"同时删除实例数据（PVC），不可恢复"`
}

type DeleteRes struct{}

// ListOrphanPVCsReq 列出用户可复用的孤立 PVC
type ListOrphanPVCsReq struct {
	g.Meta `path:"/orphan-pvcs" method:"get" tags:"Instances" summary:"列出可复用的历史数据" dc:"列出用户命名空间下未被任何实例绑定的 PVC。"`
}

type OrphanPVC struct {
	Name         string `json:"name" dc:"PVC 名称"`
	InstanceName string `json:"instanceName" dc:"原实例名称"`
	Size         string `json:"size" dc:"存储大小"`
	CreatedAt    string `json:"createdAt" dc:"创建时间"`
}

type ListOrphanPVCsRes struct {
	List []OrphanPVC `json:"list" dc:"可复用的 PVC 列表"`
}

// DeletePVCReq 删除孤立 PVC
type DeletePVCReq struct {
	g.Meta `path:"/delete-pvc" method:"post" tags:"Instances" summary:"删除历史数据" dc:"删除用户命名空间下的孤立 PVC，不可恢复。"`

	Name string `json:"name" v:"required" dc:"PVC 名称"`
}

type DeletePVCRes struct{}

// StartReq 启动实例
type StartReq struct {
	g.Meta `path:"/start" method:"post" tags:"Instances" summary:"启动实例" dc:"启动已停止的实例。"`

	Id int64 `json:"id" v:"required|min:1" dc:"实例ID"`
}

type StartRes struct{}

// StopReq 停止实例
type StopReq struct {
	g.Meta `path:"/stop" method:"post" tags:"Instances" summary:"停止实例" dc:"停止运行中的实例，保留存储。"`

	Id int64 `json:"id" v:"required|min:1" dc:"实例ID"`
}

type StopRes struct{}

// RestartReq 重启实例
type RestartReq struct {
	g.Meta `path:"/restart" method:"post" tags:"Instances" summary:"重启实例" dc:"重启实例。"`

	Id int64 `json:"id" v:"required|min:1" dc:"实例ID"`
}

type RestartRes struct{}

// GetStatusReq 获取实例实时状态
type GetStatusReq struct {
	g.Meta `path:"/status" method:"get" tags:"Instances" summary:"获取实例状态" dc:"获取实例的实时 K8s 状态。"`

	Id int64 `json:"id" v:"required|min:1" dc:"实例ID"`
}

type GetStatusRes struct {
	Status   string `json:"status" dc:"实例状态"`
	PodPhase string `json:"podPhase" dc:"Pod 阶段"`
	Ready    bool   `json:"ready" dc:"是否就绪"`
}

// GetQuotaReq 获取用户配额
type GetQuotaReq struct {
	g.Meta `path:"/quota" method:"get" tags:"Instances" summary:"获取用户配额" dc:"获取当前用户的配额和使用量。"`
}

type GetQuotaRes struct {
	Quota *QuotaItem `json:"quota" dc:"配额信息"`
}
