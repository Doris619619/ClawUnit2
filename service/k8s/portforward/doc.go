// Package portforward 在 out-of-cluster 模式下建立到 Pod 的 SPDY tunnel。
//
// 公开函数列表：
//
//	GetLocal           获取/建立到 Pod 的本地转发端口（in-cluster 时返回 0）
//	Close              关闭单个 Pod 的 port-forward
//	CloseForInstance   按 instance-id 查找 Pod 并关闭其 port-forward
//
// # 使用场景
//
// ClawUnit 支持两种部署模式（详见 internal/service/k8s.Initialize）：
//
//   - in-cluster：直接用 Pod IP:18789 连 OpenClaw，不需要 port-forward。
//     GetLocal 直接返回 (0, nil)，调用方应该回退到 Pod IP。
//   - out-of-cluster：本地开发，没有 Pod 网络，需要通过 K8s API server
//     建 SPDY tunnel 把 Pod 的 18789 转到 127.0.0.1:<random>。
//
// 调用方都按 GetLocal → 失败回退 → Pod IP 的顺序处理，proxy.WsChat
// 和 proxy.LatestMedia 都是这个模式。
//
// # 连接缓存
//
// port-forward 建立成本不低（一次完整的 SPDY upgrade 握手），所以
// 同一个 Pod 的 forward 会缓存复用。每次 GetLocal 都会先 net.Dial
// 探测缓存的本地端口是否还活着 —— 如果 Pod 重启或者网络断了，
// 旧 entry 会被清理然后建新连接。
//
// 缓存是包级的 sync.Map（其实用普通 map + Mutex），key 是
// "namespace/podName"。
//
// # int32 端口的安全转换
//
// net.Listen("tcp", "127.0.0.1:0") 分配的端口必然是 0-65535 的 16-bit
// 无符号整数，永远适合 int32，所以 TCPAddr.Port → int32 的转换是
// 安全的。这里用 //nolint:forcetypeassert,gosec 注释豁免 lint，比加
// 一堆运行时 check 更简洁 —— 标准库已经保证了类型不变量。
//
// # in-cluster 模式的判断
//
// GetLocal 通过检查 c.Config.Host == "" 来识别 in-cluster 模式。这是
// client-go 的实现细节：rest.InClusterConfig() 返回的 Config.Host 会
// 被 set 成 https://kubernetes.default.svc，但 BuildConfigFromFlags
// 返回的 Config.Host 是 kubeconfig 里的实际 server URL —— 当前判断
// 实际上是反的（in-cluster 应该是非空），如果发现 in-cluster 也走
// port-forward 路径需要重新审视这里。
package portforward
