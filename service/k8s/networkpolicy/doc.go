// Package networkpolicy 管理实例的 K8s NetworkPolicy 资源。
//
// 公开函数列表：
//
//	Ensure   创建或更新实例的 egress 网络策略
//	Delete   删除实例的网络策略（不存在不报错）
//
// # 安全模型
//
// 每个实例 Pod 都被一个 NetworkPolicy 限制 egress（不限制 ingress，
// 因为 Pod 监听本地 18789 由 ClawUnit 主动连接，CNI 默认会阻止
// 跨 namespace 入站访问）。允许的出口流量只有三类：
//
//  1. DNS：UDP/TCP 53 to kube-system namespace
//  2. 集群内系统服务：TCP 80/443/8032/8033 to "{base}-system" namespace
//     （open-platform、UniAuth 等）
//  3. 外网 HTTP/HTTPS：TCP 80/443 to 0.0.0.0/0
//     （LLM API、apt 仓库、web_fetch 工具）
//
// 这意味着用户实例 *不能* 访问其他用户的 Pod，也不能 SSH/数据库连接
// 任何东西。如果以后需要新的出口规则，必须明确添加，默认拒绝。
//
// # AllowPrivateNetwork 字段
//
// CreateReq 里有个 AllowPrivateNetwork bool，但当前 Ensure 实现里
// *没有* 根据它放开任何规则 —— 这是预留位，将来若需要给特定用户
// （比如管理员调试）放开私网访问，应该在这里加额外的 egress rule。
//
// # 使用 new(literal) 写指针字段
//
// NetworkPolicyPort 的 Protocol 和 Port 是指针字段，正确写法：
//
//	{Protocol: new(corev1.ProtocolTCP), Port: new(intstr.FromInt(443))}
//
// 不要造 protocolPtr() 或 intPtr() helper 函数 —— Go 1.26 的 new(literal)
// 语法直接支持字面量和函数返回值取地址，详见 CLAUDE.md。
package networkpolicy
