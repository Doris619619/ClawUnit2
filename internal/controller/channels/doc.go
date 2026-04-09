// Package channels 实现渠道插件管理 HTTP API。
//
// 实现 api/channels/v1 中的 IChannelsV1 interface（Install、Uninstall、
// RestartGateway、Status 四个方法）。
//
// LoginSSE 是个例外：扫码登录需要 SSE 流式返回二维码和实时状态，
// 无法用 controller.Bind 模式表达，所以写成包级函数 LoginSSE，由
// internal/cmd/cmd.go 用原生 ghttp.Request handler 注册到
// /api/channels/v1/login。包级 helper（getInstancePod、sseStreamWriter）
// 也放在同名文件 channels.go 里。
//
// 安装/卸载流程：
//
//  1. 在 Pod 内执行 npm install 把插件包写入用户 PVC 的 extensions/ 目录
//     （借助 internal/service/k8s/exec.InPodStream 流式获取输出）。
//  2. 安装成功后调 internal/service/k8s/configmap.PatchPlugin 更新
//     ConfigMap 的 plugins.allow 和 plugins.entries。
//  3. 用户调用 RestartGateway 重启 Gateway 进程加载新插件 ——
//     不会重启容器，已建立的 WebSocket session 不受影响。
package channels
