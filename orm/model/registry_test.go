package model

import (
	"database/sql"
	"github.com/coderi421/kyuu/orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestWithTableName(t *testing.T) {
	testCases := []struct {
		name          string
		val           any
		opt           Option
		wantTableName string
		wantErr       error
	}{
		{
			name:          "empty string",
			val:           &TestModel{},
			opt:           WithTableName(""),
			wantTableName: "",
		},
		{
			name:          "table name",
			val:           &TestModel{},
			opt:           WithTableName("test_model_t"),
			wantTableName: "test_model_t",
		},
	}

	r := NewRegistry()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Register(tc.val, tc.opt)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantTableName, m.TableName)
		})
	}
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}

func TestWithColumnName(t *testing.T) {
	testCases := []struct {
		name        string
		val         any
		opt         Option
		field       string
		wantColName string
		wantErr     error
	}{
		{
			name:        "new name",
			val:         &TestModel{},
			opt:         WithColumnName("FirstName", "first_name_new"),
			field:       "FirstName",
			wantColName: "first_name_new",
		},
		{
			name:        "empty new name",
			val:         &TestModel{},
			opt:         WithColumnName("FirstName", ""),
			field:       "FirstName",
			wantColName: "",
		},
		{
			// 不存在的字段
			name:    "invalid Field name",
			val:     &TestModel{},
			opt:     WithColumnName("FirstNameXXX", "first_name"),
			field:   "FirstNameXXX",
			wantErr: errs.NewErrUnknownField("FirstNameXXX"),
		},
	}

	r := NewRegistry().(*registry)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Register(tc.val, tc.opt)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			fd := m.FieldMap[tc.field]
			assert.Equal(t, tc.wantColName, fd.ColName)
		})
	}
}
func TestRegistry_get(t *testing.T) {
	testCases := []struct {
		name      string
		val       any
		wantModel *Model
		wantErr   error
	}{
		{
			name:    "test model",
			val:     TestModel{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name: "pointer",
			val:  &TestModel{},
			wantModel: &Model{
				TableName: "test_model",
				FieldMap: map[string]*Field{
					"Id": {
						ColName: "id",
						GoName:  "Id",
						Type:    reflect.TypeOf(int64(0)),
						Offset:  0,
					},
					"FirstName": {
						ColName: "first_name",
						GoName:  "FirstName",
						Type:    reflect.TypeOf(""),
						Offset:  8,
					},
					"Age": {
						ColName: "age",
						GoName:  "Age",
						Type:    reflect.TypeOf(int8(0)),
						Offset:  24,
					},
					"LastName": {
						ColName: "last_name",
						GoName:  "LastName",
						Type:    reflect.TypeOf(&sql.NullString{}),
						Offset:  32,
					},
				},
				ColumnMap: map[string]*Field{
					"id": {
						ColName: "id",
						GoName:  "Id",
						Type:    reflect.TypeOf(int64(0)),
						Offset:  0,
					},
					"first_name": {
						ColName: "first_name",
						GoName:  "FirstName",
						Type:    reflect.TypeOf(""),
						Offset:  8,
					},
					"age": {
						ColName: "age",
						GoName:  "Age",
						Type:    reflect.TypeOf(int8(0)),
						Offset:  24,
					},
					"last_name": {
						ColName: "last_name",
						GoName:  "LastName",
						Type:    reflect.TypeOf(&sql.NullString{}),
						Offset:  32,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "multiple pointer",
			val: func() any {
				val := &TestModel{}
				return &val
			}(),
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "map",
			val:     map[string]string{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "slice",
			val:     []int{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "basic type",
			val:     0,
			wantErr: errs.ErrPointerOnly,
		},

		// 标签相关测试用例
		{
			name: "column tag",
			val: func() any {
				// 我们把测试结构体定义在方法内部，防止被其它用例访问
				type ColumnTag struct {
					ID uint64 `orm:"column=id"`
				}
				return &ColumnTag{}
			}(),
			wantModel: &Model{
				TableName: "column_tag",
				FieldMap: map[string]*Field{
					"ID": {
						ColName: "id",
						GoName:  "ID",
						Type:    reflect.TypeOf(uint64(0)),
					},
				},
				ColumnMap: map[string]*Field{
					"id": {
						ColName: "id",
						GoName:  "ID",
						Type:    reflect.TypeOf(uint64(0)),
					},
				},
			},
		},
		{
			name: "empty column",
			val: func() any {
				type EmptyColumn struct {
					FirstName string `orm:"column="`
				}
				return &EmptyColumn{}
			}(),
			wantModel: &Model{
				TableName: "empty_column",
				FieldMap: map[string]*Field{
					"FirstName": {
						ColName: "first_name",
						GoName:  "FirstName",
						Type:    reflect.TypeOf(""),
					},
				},
				ColumnMap: map[string]*Field{
					"first_name": {
						ColName: "first_name",
						GoName:  "FirstName",
						Type:    reflect.TypeOf(""),
					},
				},
			},
		},
		{
			name: "invalid tag",
			val: func() any {
				type InvalidTag struct {
					FirstName string `orm:"column"`
				}
				return &InvalidTag{}
			}(),
			wantErr: errs.NewErrInvalidTagContent("column"),
		},
		{
			name: "ignore tag",
			val: func() any {
				type IgnoreTag struct {
					FirstName string `orm:"aaa=aaa"`
				}
				return &IgnoreTag{}
			}(),
			wantModel: &Model{
				TableName: "ignore_tag",
				FieldMap: map[string]*Field{
					"FirstName": {
						ColName: "first_name",
						GoName:  "FirstName",
						Type:    reflect.TypeOf(""),
					},
				},
				ColumnMap: map[string]*Field{
					"first_name": {
						ColName: "first_name",
						GoName:  "FirstName",
						Type:    reflect.TypeOf(""),
					},
				},
			},
		},
		// 利用接口自定义模型信息
		{
			name: "table name",
			val:  &CustomTableName{},
			wantModel: &Model{
				TableName: "custom_table_name_t",
				FieldMap: map[string]*Field{
					"Name": {
						ColName: "name",
						GoName:  "Name",
						Type:    reflect.TypeOf(""),
					},
				},
				ColumnMap: map[string]*Field{
					"name": {
						ColName: "name",
						GoName:  "Name",
						Type:    reflect.TypeOf(""),
					},
				},
			},
		},
		{
			name: "table name ptr",
			val:  &CustomTableNamePtr{},
			wantModel: &Model{
				TableName: "custom_table_name_ptr_t",
				FieldMap: map[string]*Field{
					"Name": {
						ColName: "name",
						GoName:  "Name",
						Type:    reflect.TypeOf(""),
					},
				},
				ColumnMap: map[string]*Field{
					"name": {
						ColName: "name",
						GoName:  "Name",
						Type:    reflect.TypeOf(""),
					},
				},
			},
		},
		{
			name: "empty table name",
			val:  &EmptyTableName{},
			wantModel: &Model{
				TableName: "empty_table_name",
				FieldMap: map[string]*Field{
					"Name": {
						ColName: "name",
						GoName:  "Name",
						Type:    reflect.TypeOf(""),
					},
				},
				ColumnMap: map[string]*Field{
					"name": {
						ColName: "name",
						GoName:  "Name",
						Type:    reflect.TypeOf(""),
					},
				},
			},
		},
	}

	r := &registry{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Get(tc.val)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantModel, m)
		})
	}
}

func Test_underscoreName(t *testing.T) {
	testCases := []struct {
		name    string
		srcStr  string
		wantStr string
	}{
		// 我们这些用例就是为了确保
		// 在忘记 underscoreName 的行为特性之后
		// 可以从这里找回来
		// 比如说过了一段时间之后
		// 忘记了 ID 不能转化为 id
		// 那么这个测试能帮我们确定 ID 只能转化为 i_d
		{
			name:    "upper cases",
			srcStr:  "ID",
			wantStr: "i_d",
		}, {
			name:    "use number",
			srcStr:  "Table1Name",
			wantStr: "table1_name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := underscoreName(tc.srcStr)
			assert.Equal(t, tc.wantStr, res)
		})
	}
}

type CustomTableName struct {
	Name string
}

func (c CustomTableName) TableName() string {
	return "custom_table_name_t"
}

type CustomTableNamePtr struct {
	Name string
}

func (c *CustomTableNamePtr) TableName() string {
	return "custom_table_name_ptr_t"
}

type EmptyTableName struct {
	Name string
}

func (c *EmptyTableName) TableName() string {
	return ""
}
