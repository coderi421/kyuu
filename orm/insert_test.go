package orm

import (
	"database/sql"
	"github.com/coderi421/kyuu/orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInserter_Build(t *testing.T) {
	db := memoryDB(t)

	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			// 不提供数据
			name:    "no value",
			q:       NewInserter[TestModel](db).Values(),
			wantErr: errs.ErrInsertZeroRow,
		},
		{
			name: "single values",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Zheng",
					Age:       18,
					LastName:  &sql.NullString{String: "Tianyi", Valid: true},
				}),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model` (`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?);",
				Args: []any{int64(1), "Zheng", int8(18), &sql.NullString{String: "Tianyi", Valid: true}},
			},
			wantErr: nil,
		},
		{
			name: "miltiple values",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Zheng",
					Age:       18,
					LastName:  &sql.NullString{String: "Tianyi", Valid: true},
				},
				&TestModel{
					Id:        2,
					FirstName: "Tom",
					Age:       16,
					LastName:  &sql.NullString{String: "Jerry", Valid: true},
				}),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model` (`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?),(?,?,?,?);",
				Args: []any{int64(1), "Zheng", int8(18), &sql.NullString{String: "Tianyi", Valid: true}, int64(2), "Tom", int8(16), &sql.NullString{String: "Jerry", Valid: true}},
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}

}