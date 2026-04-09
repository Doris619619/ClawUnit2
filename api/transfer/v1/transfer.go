package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ExportReq 导出 OpenClaw 配置
type ExportReq struct {
	g.Meta `path:"/export" method:"get" tags:"Transfer" summary:"导出配置" dc:"导出 OpenClaw 实例配置为 tar.gz。"`

	InstanceId int64 `json:"instanceId" v:"required|min:1" dc:"实例ID"`
}

type ExportRes struct {
	// 响应为二进制流，不走统一响应格式
}

// ImportReq 导入 OpenClaw 配置
type ImportReq struct {
	g.Meta `path:"/import" method:"post" tags:"Transfer" summary:"导入配置" dc:"导入 OpenClaw 配置 tar.gz 到实例。"`

	InstanceId int64 `json:"instanceId" v:"required|min:1" dc:"实例ID"`
	// 文件通过 multipart/form-data 上传
}

type ImportRes struct{}
