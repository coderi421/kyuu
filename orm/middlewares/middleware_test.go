package middlewares

import (
	"context"
	"database/sql"
	"github.com/coderi421/kyuu/orm"
	"github.com/coderi421/kyuu/orm/middlewares/querylog"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestNewMiddlewareBuilder(t *testing.T) {
	var query string
	var args []any

	customLogFunc := func(q string, as []any) {
		query = q
		args = as
		log.Printf("sql: %s, args: %v", query, args)
	}

	m := (&querylog.MiddlewareBuilder{}).LogFunc(customLogFunc)

	db, err := orm.Open("sqlite3",
		"file:test.db?cache=shared&mode=memory",
		orm.DBWithMiddlewares(m.Build()))
	require.NoError(t, err)
	_, _ = orm.NewSelector[TestModel](db).Where(orm.C("Id").EQ(10)).Get(context.Background())
	assert.Equal(t, "SELECT * FROM `test_model` WHERE `id` = ?;", query)
	assert.Equal(t, []any{10}, args)

	orm.NewInserter[TestModel](db).Values(&TestModel{Id: 18}).Exec(context.Background())
	assert.Equal(t, "INSERT INTO `test_model` (`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?);", query)
	assert.Equal(t, []any{int64(18), "", int8(0), (*sql.NullString)(nil)}, args)
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
