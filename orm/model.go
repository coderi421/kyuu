package orm

import "github.com/coderi421/kyuu/orm/internal/errs"

// ModelOpt is a function type that modifies a Model.
type ModelOpt func(model *Model) error

// Model 结构体映射db后的结构
type Model struct {
	// tableName 结构体对应的表名
	tableName string
	fieldMap  map[string]*field
}

// WithTableName is a ModelOpt function that sets the table name for a Model.
func ModelWithTableName(tableName string) ModelOpt {
	return func(model *Model) error {
		model.tableName = tableName
		return nil
	}
}

// ModelWithColumnName is a function that returns a ModelOpt function, which can be used to set the column name for a specific field in a model.
func ModelWithColumnName(field, columnName string) ModelOpt {
	return func(model *Model) error {
		// Check if the field exists in the model's field map
		fd, ok := model.fieldMap[field]
		if !ok {
			// Return an error if the field is unknown
			return errs.NewErrUnknownField(field)
		}

		// Set the column name for the field
		fd.colName = columnName
		return nil
	}
}

// field 字段相关的属性
type field struct {
	colName string
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
