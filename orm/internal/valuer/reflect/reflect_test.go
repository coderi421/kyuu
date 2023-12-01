package reflect

import (
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/coderi421/kyuu/orm/internal/errs"
	"github.com/coderi421/kyuu/orm/internal/test"
	"github.com/coderi421/kyuu/orm/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_reflectValue_SetColumns(t *testing.T) {
	testCases := []struct {
		name       string
		dbMockDate map[string][]byte
		val        *test.SimpleStruct
		wantVal    *test.SimpleStruct
		wantErr    error
	}{
		{
			name: "normal value",
			dbMockDate: map[string][]byte{
				"id":               []byte("1"),
				"bool":             []byte("true"),
				"bool_ptr":         []byte("false"),
				"int":              []byte("12"),
				"int_ptr":          []byte("13"),
				"int8":             []byte("8"),
				"int8_ptr":         []byte("-8"),
				"int16":            []byte("16"),
				"int16_ptr":        []byte("-16"),
				"int32":            []byte("32"),
				"int32_ptr":        []byte("-32"),
				"int64":            []byte("64"),
				"int64_ptr":        []byte("-64"),
				"uint":             []byte("14"),
				"uint_ptr":         []byte("15"),
				"uint8":            []byte("8"),
				"uint8_ptr":        []byte("18"),
				"uint16":           []byte("16"),
				"uint16_ptr":       []byte("116"),
				"uint32":           []byte("32"),
				"uint32_ptr":       []byte("132"),
				"uint64":           []byte("64"),
				"uint64_ptr":       []byte("164"),
				"float32":          []byte("3.2"),
				"float32_ptr":      []byte("-3.2"),
				"float64":          []byte("6.4"),
				"float64_ptr":      []byte("-6.4"),
				"byte":             []byte("8"),
				"byte_ptr":         []byte("18"),
				"byte_array":       []byte("hello"),
				"string":           []byte("world"),
				"null_string_ptr":  []byte("null string"),
				"null_int16_ptr":   []byte("16"),
				"null_int32_ptr":   []byte("32"),
				"null_int64_ptr":   []byte("64"),
				"null_bool_ptr":    []byte("true"),
				"null_float64_ptr": []byte("6.4"),
				"json_column":      []byte(`{"name": "Tom"}`),
			},
			val:     &test.SimpleStruct{},
			wantVal: test.NewSimpleStruct(1),
		},
		{
			name: "invalid field",
			dbMockDate: map[string][]byte{
				"invalid_column": nil,
			},
			wantErr: errs.NewErrUnknownColumn("invalid_column"),
		},
	}

	// 用于存储 go struct 和 db table 的映射信息的实例
	r := model.NewRegistry()

	// 获取 go SimpleStruct 和 db simple_struct 表的映射信息
	meta, err := r.Get(&test.SimpleStruct{})
	if err != nil {
		t.Fatal(err)
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//	使用 sqlmock 模拟数据库
			db, mock, errM := sqlmock.New()
			if errM != nil {
				t.Fatal(errM)
			}
			// 关闭数据库
			defer func() { _ = db.Close() }()

			// 将 go struct 和 映射信息传递进去，让后续的 DB 数据映射到 go struct 中
			reflectVal := NewReflectValue(testCase.val, meta)

			// 使用 sqlmock 构造测试用的 db data
			// column name
			cols := make([]string, 0, len(testCase.dbMockDate))
			// row data
			colVals := make([]driver.Value, 0, len(testCase.dbMockDate))
			// 将 mock 数据插入到对应 slice 中
			for c, v := range testCase.dbMockDate {
				cols = append(cols, c)
				colVals = append(colVals, v)
			}
			//	构造数据，当执行 sql 的时候，返回 ** 数据
			mock.ExpectQuery("SELECT *").
				WillReturnRows(sqlmock.NewRows(cols).
					AddRow(colVals...))

			// 模拟调用数据库
			rows, _ := db.Query("SELECT *")
			rows.Next()

			// 虽然没有直接操作 testCase.val， 但是由于是指针，所以在 SetColumns 方法中，
			// 将 db 值映射到了 testCase.val 中，所以实际上 testCase.val 中的值已经被改变了
			err = reflectVal.SetColumns(rows)
			if err != nil {
				assert.Equal(t, testCase.wantErr, err)
				return
			}
			if testCase.wantErr != nil {
				t.Fatalf("期望得到错误，但是并没有得到 %v", testCase.wantErr)
			}
			assert.Equal(t, testCase.wantVal, testCase.val)
		})
	}
}
