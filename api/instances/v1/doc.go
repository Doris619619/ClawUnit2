// Package v1 是实例管理 API 的 request/response 定义。
//
// 每个带 `g.Meta` tag 的 struct 都对应一个 HTTP 路由：
//
//	GetListReq         GET  /api/instances/v1/list           分页查询当前用户的实例
//	GetOneReq          GET  /api/instances/v1/detail         按 id 查实例详情
//	CreateReq          POST /api/instances/v1/create         创建实例并分配 K8s 资源
//	UpdateReq          POST /api/instances/v1/update         更新实例名称和描述
//	UpdateConfigReq    POST /api/instances/v1/config         热更新 OpenClaw 配置（不重启）
//	DeleteReq          POST /api/instances/v1/delete         删除实例及其所有 K8s 资源
//	StartReq           POST /api/instances/v1/start          启动 stopped/error 实例
//	StopReq            POST /api/instances/v1/stop           停止运行中的实例（保留 PVC）
//	RestartReq         POST /api/instances/v1/restart        重启实例（Stop → Start）
//	GetStatusReq       GET  /api/instances/v1/status         查询 K8s 实时状态
//	GetQuotaReq        GET  /api/instances/v1/quota          查询当前用户配额和使用量
//	ListOrphanPVCsReq  GET  /api/instances/v1/orphan-pvcs    列出可复用的历史数据 PVC
//	DeletePVCReq       POST /api/instances/v1/delete-pvc     删除孤立的历史数据 PVC
//
// 校验规则用 GoFrame 的 v 标签写在字段上（例如 v:"required|min:1"），
// 路由前缀和 method 写在 g.Meta 里，描述用 dc 标签 —— 这些字段都会被
// `gf gen ctrl` 反射读取并写入生成的 OpenAPI doc。
//
// # CreateReq 的双模式 API 配置
//
// ApiMode 字段决定 LLM 接入方式：
//
//   - "manual"（默认）：用户提供 ApiKey + BaseUrl + ModelID，直接打 LLM。
//   - "auto"：通过 UniAuth/Open Platform 自动分配 API Key，绑定到 QuotaPool。
//
// 两种模式下生成的 K8s Pod 环境变量不同（CUSTOM_API_KEY / CUSTOM_BASE_URL），
// 详见 internal/service/lifecycle/create.go。
//
// # ExistingPVC 字段
//
// 创建实例时 ExistingPvc 不为空表示复用历史数据 PVC（之前删除实例
// 但保留了用户数据）。新建 PVC 用 internal/service/k8s/pvc.EnsureInstance
// 完成；复用历史 PVC 时会校验 PVC label 中的 owner-upn 是否匹配。
package v1
