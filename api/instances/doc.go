// Package instances 定义实例管理 API 的 controller interface。
//
// IInstancesV1 由 `gf gen ctrl` 从 api/instances/v1/instances.go 里
// 带 `g.Meta` tag 的 request 结构自动生成，不要手工编辑这里的文件。
// 修改 v1 里的 request 之后跑一次 `gf gen ctrl` 重新生成 interface 和
// controller stub。
//
// 实现位于 internal/controller/instances，每个方法一个文件
// （instances_v1_create.go、instances_v1_get_list.go 等）。
package instances
