package orm

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/coderi421/kyuu/orm/internal/errs"
	"github.com/coderi421/kyuu/orm/internal/valuer/reflect"
	"github.com/coderi421/kyuu/orm/internal/valuer/unsafe"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSelector_Build(t *testing.T) {
	db := memoryDB(t)

	type testCase struct {
		name    string
		q       QueryBuilder
		want    *Query
		wantErr error
	}
	tests := []testCase{
		{
			name: "no from",
			q:    NewSelector[TestModel](db),
			want: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
		},
		{
			name: "with from",
			q:    NewSelector[TestModel](db).From("`test_model`"),
			want: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
		},
		{
			name: "empty from",
			q:    NewSelector[TestModel](db).From(""),
			want: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
		},
		{
			name: "with db",
			q:    NewSelector[TestModel](db).From("`test_db`.`test_model`"),
			want: &Query{
				SQL: "SELECT * FROM `test_db`.`test_model`;",
			},
		},
		{
			name: "single and simple predicate",
			q:    NewSelector[TestModel](db).From("`test_model`").Where(C("Id").EQ(1)),
			want: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `id` = ?;",
				Args: []any{1},
			},
		},
		{
			name: "multiple predicates",
			q:    NewSelector[TestModel](db).Where(C("Age").GT(11), C("Age").LT(13)),
			want: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` > ?) AND (`age` < ?);",
				Args: []any{11, 13},
			},
		},
		{
			// 使用 AND
			name: "and",
			q: NewSelector[TestModel](db).
				Where(C("Age").GT(18).And(C("Age").LT(35))),
			want: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` > ?) AND (`age` < ?);",
				Args: []any{18, 35},
			},
		},
		{
			// 使用 OR
			name: "or",
			q: NewSelector[TestModel](db).
				Where(C("Age").GT(18).Or(C("Age").LT(35))),
			want: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` > ?) OR (`age` < ?);",
				Args: []any{18, 35},
			},
		},
		{
			// 使用 NOT
			name: "not",
			q:    NewSelector[TestModel](db).Where(Not(C("Age").GT(18))),
			want: &Query{
				// NOT 前面有两个空格，因为我们没有对 NOT 进行特殊处理
				SQL:  "SELECT * FROM `test_model` WHERE  NOT (`age` > ?);",
				Args: []any{18},
			},
		},
		{
			name:    "invalid column",
			q:       NewSelector[TestModel](db).Where(Not(C("Invalid").GT(18))),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := tt.q.Build()
			assert.Equal(t, tt.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tt.want, query)
		})
	}
}

func TestSelector_Get(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = mockDB.Close()
	}()

	db, err := OpenDB(mockDB)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name     string
		query    string
		mockErr  error
		mockRows *sqlmock.Rows
		wantErr  error
		wantVal  *TestModel
	}{
		{
			// 查询返回错误
			name:    "query error",
			mockErr: errors.New("invalid query"),
			wantErr: errors.New("invalid query"),
			query:   "SELECT .*",
		},
		{
			name:     "no row",
			query:    "SELECT .*",
			mockRows: sqlmock.NewRows([]string{"id"}),
			wantErr:  ErrNoRows,
		},
		{
			// 数据库返回的列过多，struct 对应的列过少
			name:    "return too many column",
			wantErr: errs.ErrTooManyReturnedColumns,
			query:   "SELECT .*",
			mockRows: func() *sqlmock.Rows {
				res := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name", "extra_column"})
				res.AddRow([]byte("1"), []byte("Da"), []byte("18"), []byte("Ming"), []byte("nothing"))
				return res
			}(),
		},
		{
			name:  "get data",
			query: "SELECT .*",
			mockRows: func() *sqlmock.Rows {
				res := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
				res.AddRow([]byte("1"), []byte("Da"), []byte("18"), []byte("Ming"))
				return res
			}(),
			wantVal: &TestModel{
				Id:        1,
				FirstName: "Da",
				Age:       18,
				LastName: &sql.NullString{
					String: "Ming",
					Valid:  true,
				},
			},
		},
	}

	//for _, tc := range testCases {
	//
	//	exp := mock.ExpectQuery(tc.query)
	//	if tc.mockErr != nil {
	//		exp.WillReturnError(tc.mockErr)
	//	} else {
	//		exp.WillReturnRows(tc.mockRows)
	//	}
	//
	//}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			exp := mock.ExpectQuery(tc.query)
			if tc.mockErr != nil {
				exp.WillReturnError(tc.mockErr)
			} else {
				exp.WillReturnRows(tc.mockRows)
			}

			res, err1 := NewSelector[TestModel](db).Get(context.Background())
			assert.Equal(t, tc.wantErr, err1)
			if err1 != nil {
				return
			}
			assert.Equal(t, tc.wantVal, res)
		})
	}
}

// 在 orm 目录下执行
// go test -bench=BenchmarkQuerier_Get -benchmem -benchtime=10000x
// 我的输出结果
//goos: windows
//goarch: amd64
//pkg: github.com/coderi421/kyuu/orm
//cpu: 11th Gen Intel(R) Core(TM) i5-11400 @ 2.60GHz
//BenchmarkQuerier_Get/unsafe-12             10000            220244 ns/op            3399 B/op        111 allocs/op
//BenchmarkQuerier_Get/reflect-12            10000            743207 ns/op            3581 B/op        120 allocs/op
//PASS
//ok      github.com/coderi421/kyuu/orm   11.026s

func BenchmarkQuerier_Get(b *testing.B) {
	db, err := Open("sqlite3", "file:benchmark_get.db?cache=shared&mode=memory")
	if err != nil {
		b.Fatal(err)
	}
	_, err = db.db.Exec(TestModel{}.CreateSQL())
	if err != nil {
		b.Fatal(err)
	}

	res, err := db.db.Exec("INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?)", 12, "Tom", 18, "Jerry")
	if err != nil {
		b.Fatal(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		b.Fatal(err)
	}
	if affected == 0 {
		b.Fatal()
	}

	b.ResetTimer()
	b.Run("unsafe", func(b *testing.B) {
		db.valCreator = unsafe.NewUnsafeValue
		for i := 0; i < b.N; i++ {
			_, err = NewSelector[TestModel](db).Get(context.Background())
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("reflect", func(b *testing.B) {
		db.valCreator = reflect.NewReflectValue
		for i := 0; i < b.N; i++ {
			_, err = NewSelector[TestModel](db).Get(context.Background())
			if err != nil {
				b.Fatal()
			}
		}
	})
}
