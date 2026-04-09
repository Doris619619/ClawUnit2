// Package v1 是 OpenClaw 配置导入导出 API 的 request/response 定义。
//
// 路由列表：
//
//	ExportReq  GET  /api/transfer/v1/export   把指定实例的 OpenClaw 配置打包成 tar.gz
//	ImportReq  POST /api/transfer/v1/import   把 tar.gz 还原到目标实例
//
// # 流式响应
//
// Export 的响应是二进制 tar.gz 流，不走 UniResMiddleware 的 JSON 包装
// （Content-Type 设为 application/octet-stream 触发中间件的流式短路）。
//
// Import 的请求是 multipart/form-data，文件通过 r.GetUploadFile 接收。
// internal/cmd/cmd.go 把 client max body size 调到 512MB 就是为了 Import。
//
// # 包含哪些配置
//
// 导出范围是 OpenClaw Pod 中 /home/node/.openclaw 目录下的所有可移植
// 数据：openclaw.json、agents/、memory/ 等，但不包含 workspace（用户运行
// 时数据，跨实例不通用）和 .cache（Playwright 浏览器，全用户共享）。
package v1
