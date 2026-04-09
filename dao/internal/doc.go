// Package internal 存放 `gf gen dao` 自动生成的底层 DAO struct，
// 不要手工编辑这里的任何文件。
//
// 每张数据库表对应一个 XxxDao struct（例如 InstancesDao）和一个 XxxColumns
// struct（包含所有列名常量）。父包 internal/dao 通过组合这些 struct 暴露
// 全局对象 dao.Instances 等给业务代码使用，外部不应该直接 import 本包。
//
// 触发重新生成：修改 manifest/deploy/migrations/*.sql 后运行 `gf gen dao`。
package internal
