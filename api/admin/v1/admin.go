package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// AdminListInstancesReq 管理员查看全局实例列表
type AdminListInstancesReq struct {
	g.Meta `path:"/instances/list" method:"get" tags:"Admin" summary:"全局实例列表" dc:"管理员查看所有用户的实例。"`

	OwnerUpn string `json:"ownerUpn" dc:"按用户过滤"`
	Status   string `json:"status" v:"in:creating,running,stopped,error,deleting" dc:"按状态过滤"`
	Page     int    `json:"page" d:"1" v:"min:1" dc:"页码"`
	PageSize int    `json:"pageSize" d:"20" v:"min:1|max:200" dc:"每页数量"`
}

type AdminListInstancesRes struct {
	List  []*AdminInstanceItem `json:"list" dc:"实例列表"`
	Total int                  `json:"total" dc:"总数"`
}

// AdminInstanceItem 管理员视角的实例信息
type AdminInstanceItem struct {
	Name         string `json:"name" dc:"实例名称"`
	Status       string `json:"status" dc:"实例状态"`
	OwnerUpn     string `json:"ownerUpn" dc:"所属用户"`
	PodName      string `json:"podName" dc:"Pod 名称"`
	PodNamespace string `json:"podNamespace" dc:"Pod 命名空间"`
	CpuCores     string `json:"cpuCores" dc:"CPU 资源量"`
	MemoryGb     string `json:"memoryGb" dc:"内存资源量"`
	DiskGb       string `json:"diskGb" dc:"磁盘资源量"`
	Id           int64  `json:"id" dc:"实例ID"`
}

// AdminUpdateQuotaReq 管理员设置用户配额（前期只管理实例数量上限）
type AdminUpdateQuotaReq struct {
	g.Meta `path:"/quotas/update" method:"post" tags:"Admin" summary:"设置用户配额" dc:"设置或更新指定用户的资源配额。"`

	MaxInstances *int32 `json:"maxInstances" v:"min:0" dc:"最大实例数"`
	OwnerUpn     string `json:"ownerUpn" v:"required" dc:"用户 UPN"`
}

type AdminUpdateQuotaRes struct{}

// SyncStatusReq 获取同步服务状态
type SyncStatusReq struct {
	g.Meta `path:"/sync/status" method:"get" tags:"Admin" summary:"同步状态" dc:"获取 K8s 同步服务的运行状态。"`
}

type SyncStatusRes struct {
	LastRun string `json:"lastRun" dc:"上次同步时间"`
	Running bool   `json:"running" dc:"是否运行中"`
}

// ForceSyncReq 强制同步
type ForceSyncReq struct {
	g.Meta `path:"/sync/force" method:"post" tags:"Admin" summary:"强制同步" dc:"立即触发一次 K8s 状态同步。"`
}

type ForceSyncRes struct{}
