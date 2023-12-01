package model

import "reflect"

// Option is a function type that modifies a Model.
type Option func(model *Model) error

// Model 结构体映射db后的结构
type Model struct {
	// TableName 结构体对应的表名
	TableName string
	FieldMap  map[string]*Field // 结构体 属性名 attr name 为 key  ItemId
	ColumnMap map[string]*Field // DB column name 为 key    item_id
}

// Field 字段相关的属性
type Field struct {
	ColName string       // 数据库中的字段名
	GoName  string       // go struct 中的名字
	Type    reflect.Type // go 中的数据类型，转换成 reflect.Value 的时候，知道是什么类型，不然那没法转
	// Offset 相对于对象起始地址的字段偏移量
	// uintptr 这个类型的值，只是简单记录一下位置
	Offset uintptr
}

// 我们支持的全部标签上的 key 都放在这里
// 方便用户查找，和我们后期维护
const (
	tagKeyColumn = "column"
	tagORMName   = "orm"
)

// TableName 用户实现这个接口来返回自定义的表名
type TableName interface {
	TableName() string
}
