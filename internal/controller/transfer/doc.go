// Package transfer 实现 OpenClaw 配置导入导出 HTTP API。
//
// 实现 api/transfer/v1 中的 ITransferV1 interface。
//
// Export 通过 K8s exec 在 Pod 内 tar 打包 /home/node/.openclaw 目录
// 然后流式吐到 HTTP 响应；Import 反向操作，从上传文件解包到 Pod。
// 所有响应都是二进制流，注意要在 handler 里设置 Content-Type 让
// UniResMiddleware 跳过 JSON 包装。
package transfer
