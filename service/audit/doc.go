// Package audit 提供审计日志写入接口。
//
// 公开函数列表：
//
//	Log   把一次操作记录写入 audit_events 表
//
// # 设计原则
//
// 审计是"尽力而为"操作 —— 写失败只记 error log，绝不回返调用方，
// 这样审计模块挂掉不会阻塞业务请求。但 caller 应该在所有改变状态
// 的操作完成 *之后* 调用 Log，确保 details 反映的是最终结果。
//
// # 字段约定
//
//	actorUpn      操作人的 UPN（管理员代用户操作时填管理员）
//	action        动词，如 "instance.create" / "instance.delete" / "channels.install"
//	resourceType  资源类型，如 "instance" / "skill" / "channel"
//	resourceID    资源 ID（int64，对应业务表的主键）
//	details       g.Map，存任意上下文（请求参数、关键决策等），自动序列化为 JSON
//
// # 不要在 details 里塞敏感数据
//
// API key 原文、密码、access token 都不能写入 details —— audit_events
// 表是给运维和合规审查看的，会长期保留。写之前过一遍：如果泄露这条
// 记录会带来安全风险吗？如果会，去掉那个字段。
//
// # 与日志的区别
//
//   - g.Log() 是运行时观测，目标是排查问题，可能被采样/压缩/删除
//   - audit.Log 是合规审计，目标是事后追溯，必须持久化到 DB
//
// 不要混用 —— 运行调试用 g.Log()，重要的状态变化用 audit.Log。
package audit
