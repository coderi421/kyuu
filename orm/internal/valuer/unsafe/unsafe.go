package unsafe

import (
	"database/sql"
	"github.com/coderi421/kyuu/orm/internal/errs"
	"github.com/coderi421/kyuu/orm/internal/valuer"
	"github.com/coderi421/kyuu/orm/model"
	"reflect"
	"unsafe"
)

type unsafeValue struct {
	addr unsafe.Pointer // 使用 unsafe Pointer 而不是 uintptr 是因为 gc 后 uintptr 会发生变化
	meta *model.Model
}

var _ valuer.Creator = NewUnsafeValue

func NewUnsafeValue(val any, meta *model.Model) valuer.Value {
	return &unsafeValue{
		// 使用 unsafe Pointer 而不是 uintptr 是因为 gc 后 uintptr 会发生变化
		addr: unsafe.Pointer(reflect.ValueOf(val).Pointer()),
		meta: meta,
	}
}

func (u unsafeValue) SetColumns(rows *sql.Rows) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	if len(columns) > len(u.meta.ColumnMap) {
		return errs.ErrTooManyReturnedColumns
	}

	colValues := make([]any, len(columns))
	for i, column := range columns {
		cm, ok := u.meta.ColumnMap[column]
		if !ok {
			return errs.NewErrUnknownColumn(column)
		}
		ptr := unsafe.Pointer(uintptr(u.addr) + cm.Offset)
		val := reflect.NewAt(cm.Type, ptr)
		colValues[i] = val.Interface()
	}

	return rows.Scan(colValues...)
}
