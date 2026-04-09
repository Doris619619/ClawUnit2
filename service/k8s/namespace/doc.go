// Package namespace 管理用户级 K8s namespace。
//
// 公开函数列表：
//
//	Ensure   确保用户 namespace 存在，不存在则创建
//
// # 一个用户一个 namespace
//
// ClawUnit 把每个用户的所有资源放在独立 namespace 里：
//
//	namespace = "{base}-user-{upn_hash8}"   // base 来自 k8s.namespace 配置
//
// 用户 UPN 经过 SHA256 取前 4 字节得到 8 个 hex 字符的短标识，
// 拼接后再走 sanitizeK8sName 保证符合 DNS-1123 label 规则。
//
// 这种隔离方式的好处：
//
//   - 资源 quota / limit range 可以按 namespace 设
//   - 删除用户时只需要删一个 namespace 即可清空所有资源
//   - NetworkPolicy 跨 namespace 默认不通，自然就有用户级隔离
//
// 代价：每个新用户首次创建实例时多一次 Namespace 创建 API 调用。
// Ensure 是幂等的，已存在直接 return，所以不会重复创建。
//
// # 不会自动删除 namespace
//
// 删除用户最后一个实例时不会顺便删除 namespace（也不存在删除 namespace
// 的接口）。namespace 是空的也无所谓 —— 它本身不消耗资源，里面残留的
// PVC 才是关键。
package namespace
