package orm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeleter_Build(t *testing.T) {
	db, err := NewDB()
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name      string
		builder   QueryBuilder
		wantErr   error
		wantQuery *Query
	}{
		{
			name:    "no where",
			builder: NewDeleter[TestModel](db).From("`test_model`"),
			wantQuery: &Query{
				SQL: "DELETE FROM `test_model`;",
			},
		}, {
			name:    "where",
			builder: NewDeleter[TestModel](db).Where(C("Id").EQ(16)),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE `id` = ?;",
				Args: []any{16},
			},
		}, {
			name:    "from",
			builder: NewDeleter[TestModel](db).From("`test_model`").Where(C("Id").EQ(16)),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE `id` = ?;",
				Args: []any{16},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.builder.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})

	}
}
