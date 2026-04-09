// Package entity 定义"Entity"结构体，每张数据库表的 typed 容器。
//
// 文件由 `gf gen dao` 自动生成，不要手工编辑。
//
// 每个 struct 字段都有具体 Go 类型（int32 / string / *gtime.Time 等），
// 字段映射通过 `orm:"column_name"` tag 告诉 GoFrame ORM。这是用来 .Scan
// 装载查询结果的容器：
//
//	var inst entity.Instances
//	err := dao.Instances.Ctx(ctx).Where("id", id).Scan(&inst)
//	fmt.Println(inst.Status, inst.PodName)
//
// # 不要用 entity 做写操作
//
// 写操作要用 internal/model/do 里的同名 DO struct，因为 entity 的零值会
// 被序列化到 SQL（例如 entity.Instances{}.Status 是空字符串，会导致
// .Update 把 status 列改成空），而 do.Instances{}.Status 是 nil，
// 不会出现在 SQL 里。
//
// # 字段类型映射
//
// 如果生成的字段类型不符合预期（默认 int4 → int），改 hack/config.yaml
// 的 typeMapping 然后重新跑 `gf gen dao`。当前 ClawUnit 把所有 int4 都
// 映射成 Go int32 以保持代码风格一致。
package entity
