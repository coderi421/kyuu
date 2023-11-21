package orm

import (
	"github.com/coderi421/kyuu/orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSelector_Build(t *testing.T) {
	db, err := NewDB()
	if err != nil {
		t.Fatal(err)
	}

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
			q:    NewSelector[TestModel](db).From("`test_model`").Where(C("Age").GT(11), C("Age").LT(13)),
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
