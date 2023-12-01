package valuer

import (
	"database/sql"
	"github.com/coderi421/kyuu/orm/internal/errs"
	"github.com/coderi421/kyuu/orm/model"
	"reflect"
)

// reflectValue 基于反射的 Value
type reflectValue struct {
	val  reflect.Value
	meta *model.Model
}

var _ Creator = NewReflectValue

// NewReflectValue 返回一个封装好的，基于反射实现的 Value
// 输入 val 必须是一个指向结构体实例的指针，而不能是任何其它类型
func NewReflectValue(val interface{}, meta *model.Model) Value {
	return &reflectValue{
		val:  reflect.ValueOf(val).Elem(),
		meta: meta,
	}
}

// SetColumns 将数据库中的数据设置到对应的 struct 上
// SetColumns sets the values from the database to the corresponding struct.
func (r reflectValue) SetColumns(rows *sql.Rows) error {
	// Get the column names from the rows
	columnNames, err := rows.Columns()
	if err != nil {
		return err
	}

	// Check if the number of column names is greater than the number of fields in the struct
	if len(columnNames) > len(r.meta.FieldMap) {
		return errs.ErrTooManyReturnedColumns
	}

	// colValues 和 colEleValues 实质上最终都指向同一个对象
	// Create slices to hold the column values and element values
	colValues := make([]any, len(columnNames))
	colEleValues := make([]reflect.Value, len(columnNames))

	// 将 column Name 设置到 ColumnMap 中
	// Map the column names to the field names
	for i, name := range columnNames {
		// Get the field corresponding to the column name
		field, ok := r.meta.ColumnMap[name]
		if !ok {
			return errs.NewErrUnknownColumn(name)
		}

		// 构建出新的 reflect.Value struct
		// Create a new reflect.Value struct
		value := reflect.New(field.Type)

		// 实际上 colValues 和 colEleValues 存的都是指针，
		// 修改 colValues 切片中的元素时，colEleValues 切片中的相应元素也会实时获取到变化后的值。这是因为它们指向相同的内存地址。
		// 为了映射
		// Set the element value in both slices
		colValues[i] = value.Interface()
		colEleValues[i] = value.Elem()
	}

	// 这里使用 colValues 而不是 colEleValues 是因为 scan 方法接收的是 []any 参数 而不是 []reflect.Value
	// Scan the values from the rows into the colValues slice
	if err = rows.Scan(colValues...); err != nil {
		return err
	}

	// 最终 通过 r.val.FieldByName 找到结构体中的字段，将 colEleValues 中的值，赋值给这个字段
	// Set the element values to the corresponding struct fields
	for i, c := range columnNames {
		// 找到 db column name 对应的映射信息
		cm := r.meta.ColumnMap[c]
		// 通过 映射信息中的 goName 找到接收数据 Struct 中的对应字段
		fd := r.val.FieldByName(cm.GoName)
		fd.Set(colEleValues[i])
	}
	return nil
}
