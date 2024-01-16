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
		{
			// 指定列
			name: "specify columns",
			q: NewInserter[TestModel](db).Columns("FirstName", "LastName").Values(
				&TestModel{
					Id:        1,
					FirstName: "Zheng",
					Age:       18,
					LastName:  &sql.NullString{String: "Tianyi", Valid: true},
				}),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model` (`first_name`,`last_name`) VALUES (?,?);",
				Args: []any{"Zheng", &sql.NullString{String: "Tianyi", Valid: true}},
			},
		},
		{
			// 指定 非法列
			name: "invalid columns",
			q: NewInserter[TestModel](db).Columns("FirstName", "Invalid").Values(
				&TestModel{
					Id:        1,
					FirstName: "Zheng",
					Age:       18,
					LastName:  &sql.NullString{String: "Tianyi", Valid: true},
				}),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			// upset
			name: "upsert",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Zheng",
					Age:       18,
					LastName:  &sql.NullString{String: "Tianyi", Valid: true},
				}).OnDuplicateKey().Update(Assign("FirstName", "Z")),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model` (`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE `first_name`=?;",
				Args: []any{int64(1), "Zheng", int8(18), &sql.NullString{String: "Tianyi", Valid: true}, "Z"},
			},
		},
		{
			name: "upsert invalid column",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Zheng",
					Age:       18,
					LastName:  &sql.NullString{String: "Tianyi", Valid: true},
				}).OnDuplicateKey().Update(Assign("Invalid", "zheng")),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			// upset
			name: "upsert use insert value",
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
				}).OnDuplicateKey().Update(C("FirstName"), C("LastName")),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model` (`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?),(?,?,?,?) ON DUPLICATE KEY UPDATE `first_name`=VALUES(`first_name`),`last_name`=VALUES(`last_name`);",
				Args: []any{int64(1), "Zheng", int8(18), &sql.NullString{String: "Tianyi", Valid: true}, int64(2), "Tom", int8(16), &sql.NullString{String: "Jerry", Valid: true}},
			},
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

func TestUpsert_SQLite3_Build(t *testing.T) {
	db := memoryDB(t, DBWithDialect(SQLite3))
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "upsert",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Zheng",
					Age:       18,
					LastName:  &sql.NullString{String: "Tianyi", Valid: true},
				}).OnDuplicateKey().ConflictColumns("Id").Update(Assign("FirstName", "zheng")),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model` (`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?) ON CONFLICT(`id`) DO UPDATE SET `first_name`=?;",
				Args: []any{int64(1), "Zheng", int8(18), &sql.NullString{String: "Tianyi", Valid: true}, "zheng"},
			},
		},
		{
			// upsert invalid column
			name: "upsert invalid column",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Zheng",
					Age:       18,
					LastName:  &sql.NullString{String: "Tianyi", Valid: true},
				}).OnDuplicateKey().ConflictColumns("Id").
				Update(Assign("Invalid", "zheng")),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			// conflict invalid column
			name: "conflict invalid column",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Zheng",
					Age:       18,
					LastName:  &sql.NullString{String: "Tianyi", Valid: true},
				}).OnDuplicateKey().ConflictColumns("Invalid").
				Update(Assign("FirstName", "zheng")),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			// 使用原本插入的值
			name: "upsert use insert value",
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
				}).OnDuplicateKey().ConflictColumns("Id").Update(C("FirstName"), C("LastName")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model` (`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?),(?,?,?,?) ON CONFLICT(`id`) DO UPDATE SET `first_name`=excluded.`first_name`,`last_name`=excluded.`last_name`;",
				Args: []any{int64(1), "Zheng", int8(18), &sql.NullString{String: "Tianyi", Valid: true},
					int64(2), "Tom", int8(16), &sql.NullString{String: "Jerry", Valid: true}},
			},
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
