// Package dao 是 ClawUnit 的数据访问层。
//
// 文件由 `gf gen dao` 从数据库 schema 自动生成（具体来说是从
// manifest/deploy/migrations/*.sql），不要手工编辑生成的文件。修改 schema
// 之后跑一次 `gf gen dao` 重新生成。
//
// # 5 张表，5 个全局对象
//
//	dao.Instances        // instances 表，OpenClaw 实例的元数据
//	dao.UserQuotas       // user_quotas 表，每用户实例数量上限
//	dao.Skills           // skills 表，技能的元数据
//	dao.AuditEvents      // audit_events 表，审计日志
//	dao.ApiKeyProvisions // api_key_provisions 表，UniAuth API Key 的发放记录
//
// 每个对象都是包级全局变量，直接调用，不需要构造：
//
//	record, err := dao.Instances.Ctx(ctx).Where("id", id).One()
//
// # GoFrame ORM 用法约定
//
//   - Context 只在 Ctx(ctx) 一次性传入，终端方法（One/All/Count/Scan/
//     Insert/Update/Delete/InsertAndGetId）不再传 ctx。
//   - 写操作用 internal/model/do 里的 DO struct，nil 字段会被自动忽略，
//     适合做部分更新。
//   - 读操作可以 .Scan(&entity) 到 internal/model/entity 里的 typed struct。
//
// 完整示例：
//
//	// 部分更新（只改 status，其他字段不动）
//	_, err := dao.Instances.Ctx(ctx).
//	    Where("id", instanceID).
//	    Data(do.Instances{Status: "running"}).
//	    Update()
//
//	// 读到 typed struct
//	var inst entity.Instances
//	err := dao.Instances.Ctx(ctx).Where("id", instanceID).Scan(&inst)
//
// # 数据库字段类型
//
//   - 所有 int / int4 列经过 hack/config.yaml 的 typeMapping 映射到
//     Go int32（不是默认的 int）。
//   - cpu_cores / memory_gb / disk_gb 是 TEXT 列，用来直接保存 K8s 资源
//     格式字符串，例如 "1"、"2Gi"、"500Mi"。不要把它们当数字处理。
//   - 时间列是 *gtime.Time，可以为 nil（例如未启动过的实例 started_at = nil）。
//
// # 控制层 → DAO 直接调用
//
// 在 ClawUnit 里 controller 可以直接调 dao.Xxx，没必要再包一层 service。
// 只有当同一段查询逻辑出现在 ≥ 2 个地方才考虑抽到 service。这是项目
// 风格约定，详见 CLAUDE.md 的 "Layer boundaries" 一节。
package dao
