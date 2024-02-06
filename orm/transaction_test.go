package orm

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTx_Commit(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = mockDB.Close() }()

	db, err := OpenDB(mockDB)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		mock.ExpectClose()
		_ = db.Close()
	}()

	// 事务正常提交
	mock.ExpectBegin()
	mock.ExpectCommit()

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	assert.Nil(t, err)
	err = tx.Commit()
	assert.Nil(t, err)

}

func TestTx_Rollback(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = mockDB.Close() }()

	db, err := OpenDB(mockDB)
	if err != nil {
		t.Fatal(err)
	}

	// 事务回滚
	mock.ExpectBegin()
	mock.ExpectRollback()
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	assert.Nil(t, err)
	err = tx.Rollback()
	assert.Nil(t, err)
}
