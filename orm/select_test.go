package orm

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}

func TestSelector_Build(t *testing.T) {
	type testCase struct {
		name    string
		q       QueryBuilder
		want    *Query
		wantErr error
	}
	tests := []testCase{
		{
			name: "no from",
			q:    NewSelector[TestModel](),
			want: &Query{
				SQL: "SELECT * FROM `TestModel`;",
			},
		},
		{
			name: "with from",
			q:    NewSelector[TestModel]().From("`test_model`"),
			want: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
		},
		{
			name: "empty from",
			q:    NewSelector[TestModel]().From(""),
			want: &Query{
				SQL: "SELECT * FROM `TestModel`;",
			},
		},
		{
			name: "with db",
			q:    NewSelector[TestModel]().From("`test_db`.`test_model`"),
			want: &Query{
				SQL: "SELECT * FROM `test_db`.`test_model`;",
			},
		},
		{
			name: "single and simple predicate",
			q:    NewSelector[TestModel]().From("`test_model`").Where(C("Id").EQ(1)),
			want: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `Id` = ?;",
				Args: []any{1},
			},
		},
		{
			name: "multiple predicates",
			q:    NewSelector[TestModel]().From("`test_model`").Where(C("Age").GT(11), C("Age").LT(13)),
			want: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`Age` > ?) AND (`Age` < ?);",
				Args: []any{11, 13},
			},
		},
		{
			// 使用 AND
			name: "and",
			q: NewSelector[TestModel]().
				Where(C("Age").GT(18).And(C("Age").LT(35))),
			want: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE (`Age` > ?) AND (`Age` < ?);",
				Args: []any{18, 35},
			},
		},
		{
			// 使用 OR
			name: "or",
			q: NewSelector[TestModel]().
				Where(C("Age").GT(18).Or(C("Age").LT(35))),
			want: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE (`Age` > ?) OR (`Age` < ?);",
				Args: []any{18, 35},
			},
		},
		{
			// 使用 NOT
			name: "not",
			q:    NewSelector[TestModel]().Where(Not(C("Age").GT(18))),
			want: &Query{
				// NOT 前面有两个空格，因为我们没有对 NOT 进行特殊处理
				SQL:  "SELECT * FROM `TestModel` WHERE  NOT (`Age` > ?);",
				Args: []any{18},
			},
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
