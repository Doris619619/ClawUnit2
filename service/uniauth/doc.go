// Package uniauth 是 UniAuth (CUHKSZ 统一认证服务) 的客户端。
//
// 公开函数列表：
//
//	CheckPermission   检查用户对资源的权限 (sub, obj, act) 三元组
//	CreateApiKey      为用户创建 LLM API Key（绑定 Open Platform 配额池）
//	RevokeApiKey      吊销已发放的 API Key
//
// # UniAuth 在哪里
//
// UniAuth 是平台级服务，base URL 在 config.yaml 的 client.uniauth.url
// 配置。本包用 GoFrame 的 g.Client 维护一个共享 HTTP client，自动重试
// 3 次，content-type 默认 JSON。
//
// # 权限模型
//
// CheckPermission 用 Casbin 风格的 (subject, object, action) 模型：
//
//	CheckPermission(ctx, "user@cuhk.edu.cn", "clawunit", "access")  // 普通用户
//	CheckPermission(ctx, "user@cuhk.edu.cn", "clawunit", "admin")   // 管理员
//
// 这两个 obj/act 组合是 ClawUnit 用到的全部，新增权限点要先在 UniAuth
// 后台配置好规则，再加到代码里。中间件 InjectIdentity 会调第一个，
// RequireAdmin 会调第二个。
//
// # API Key 发放
//
// CreateApiKey 调 UniAuth 的 /openPlatform/apikey 端点，绑定到指定的
// 配额池（quotaPool 字符串），返回三个字段：
//
//	RawApiKey   //  完整的 API Key 文本，*只在创建响应里出现一次*
//	ApiKeyHash  //  hash 值，用于后续吊销/查询
//	MaskedKey   //  打码后的展示值（"sk-***"）
//
// ClawUnit 把 RawApiKey 直接注入 Pod 的 CUSTOM_API_KEY 环境变量，
// hash 存 instances 表的 api_key_hash 列。注意 raw key 不入库 ——
// 这意味着 Stop → Start 时无法恢复同一把 key，必须吊销旧的并创建新的
// （见 lifecycle.ensureApiKey）。
//
// # 错误语义
//
//   - CheckPermission 返回 (false, nil) 表示明确无权限，不是错误
//   - CheckPermission 返回 (_, err) 表示 UniAuth 服务异常
//   - CreateApiKey 失败应该被 caller 视为致命错误（实例创建会回滚）
//   - RevokeApiKey 失败只记 warning —— 已经在删除流程里了，不能让吊销
//     失败阻塞整个 delete
package uniauth
