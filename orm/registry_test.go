package orm

import (
	"github.com/coderi421/kyuu/orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_registry_get(t *testing.T) {
	testCases := []struct {
		name      string
		val       any
		wantModel *model
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
			wantModel: &model{
				tableName: "test_model",
				fieldMap: map[string]*field{
					"Id": {
						colName: "id",
					},
					"FirstName": {
						colName: "first_name",
					},
					"Age": {
						colName: "age",
					},
					"LastName": {
						colName: "last_name",
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
			wantModel: &model{
				tableName: "column_tag",
				fieldMap: map[string]*field{
					"ID": {
						colName: "id",
					},
				},
			},
		},
		{
			name: "empty column",
			val: func() any {
				type EmptyColumn struct {
					FirstName uint64 `orm:"column=first_name"`
				}
				return &EmptyColumn{}
			}(),
			wantModel: &model{
				tableName: "empty_column",
				fieldMap: map[string]*field{
					"FirstName": {
						colName: "first_name",
					},
				},
			},
		},
		{
			name: "invalid tag",
			val: func() any {
				type InvalidTag struct {
					FirstName uint64 `orm:"column"`
				}
				return &InvalidTag{}
			}(),
			wantErr: errs.NewErrInvalidTagContent("column"),
		},
		{
			name: "ignore tag",
			val: func() any {
				type IgnoreTag struct {
					FirstName uint64 `orm:"aaa=aaa"`
				}
				return &IgnoreTag{}
			}(),
			wantModel: &model{
				tableName: "ignore_tag",
				fieldMap: map[string]*field{
					"FirstName": {
						colName: "first_name",
					},
				},
			},
		},
		// 利用接口自定义模型信息
		{
			name: "table name",
			val:  &CustomTableName{},
			wantModel: &model{
				tableName: "custom_table_name_t",
				fieldMap: map[string]*field{
					"Name": {
						colName: "name",
					},
				},
			},
		},
		{
			name: "table name ptr",
			val:  &CustomTableNamePtr{},
			wantModel: &model{
				tableName: "custom_table_name_ptr_t",
				fieldMap: map[string]*field{
					"Name": {
						colName: "name",
					},
				},
			},
		},
		{
			name: "empty table name",
			val:  &EmptyTableName{},
			wantModel: &model{
				tableName: "empty_table_name",
				fieldMap: map[string]*field{
					"Name": {
						colName: "name",
					},
				},
			},
		},
	}

	r := &registry{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.get(tc.val)
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
