package orm

// 结构体映射db后的结构
type model struct {
	// tableName 结构体对应的表名
	tableName string
	fieldMap  map[string]*field
}

// field 字段相关的属性
type field struct {
	colName string
}
