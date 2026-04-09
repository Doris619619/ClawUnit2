// Package do 定义"Data Object"结构体，用于 GoFrame ORM 的写操作。
//
// 文件由 `gf gen dao` 自动生成，不要手工编辑。
//
// 每张数据库表对应一个 DO struct，所有字段类型都是 any（实际上是
// interface{}），这样在 .Data(do.Xxx{...}).Update() 时只有显式赋值的
// 字段会进入 SQL，其他字段保持原样不变。这是 GoFrame ORM 实现"部分
// 更新"的标准方式。
//
// 区分 DO 和 entity：
//
//   - do.Instances{Status: "running"}：写操作，nil 字段被忽略，表达
//     "只更新这一列"。
//   - entity.Instances{...}：读操作的 typed 容器，所有字段都有具体
//     Go 类型，可以 .Scan(&entity) 直接装载查询结果。
//
// 不要尝试用 do struct 接收查询结果 —— 它的字段是 any，零值不会被填充。
package do
