package k8s

// GetClient 返回全局 K8s 客户端实例。
//
// 必须在 Initialize 成功之后调用。Initialize 之前调用会返回 nil ——
// 大部分子包通过 c == nil 检测来判断是否未初始化。
func GetClient() *Client {
	return globalClient
}
