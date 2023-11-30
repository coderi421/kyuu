package model

// Option is a function type that modifies a Model.
type Option func(model *Model) error

// Model 结构体映射db后的结构
type Model struct {
	// TableName 结构体对应的表名
	TableName string
	FieldMap  map[string]*Field
}

// Field 字段相关的属性
type Field struct {
	ColName string
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
