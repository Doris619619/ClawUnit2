// Package exec 在 K8s Pod 内执行命令。
//
// 公开函数列表：
//
//	InPod        执行命令并返回 stdout/stderr 字符串（适合短输出）
//	InPodStream  执行命令并把 stdout/stderr 流式写入 io.Writer（适合长输出/二进制）
//
// 内部用 client-go 的 remotecommand SPDY executor，背后走的是 K8s API
// server 的 /exec subresource 通道。和 `kubectl exec` 等价但不依赖
// kubectl 二进制。
//
// # 何时用哪个
//
//   - 短文本输出 → InPod。例如 `ls /skills/system` 列目录、查 git revision。
//   - 长输出或二进制 → InPodStream。例如：
//   - 渠道扫码登录的 SSE 流（channels.LoginSSE）
//   - 媒体文件读取（proxy.LatestMedia 用 cat 读图片）
//   - tar 流式打包（transfer.Export）
//
// 用 InPod 接收大数据会先全部缓存到内存，可能爆 OOM；同样把
// InPodStream 写到 bytes.Buffer 再返回字符串也是反模式。
//
// # 错误处理
//
// 命令在 Pod 内非零退出会返回非 nil error，但 Result.Stdout/Stderr
// 仍然包含已经收到的输出 —— 调用方可以读 stderr 显示给用户。
// SPDY 通道断开（Pod 重启、网络抖动）也走相同的 error 路径。
//
// # 不支持 stdin
//
// 当前实现把 PodExecOptions.Stdin 写死成 false，因为目前所有调用
// 都是单向命令（执行 → 收输出）。如果以后需要交互式命令（例如远程
// shell），需要改成可选传 io.Reader，并把 SPDY stream 改成全双工。
package exec
