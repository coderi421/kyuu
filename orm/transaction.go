package orm

import (
	"context"
	"database/sql"
)

var _ Session = &Tx{}
var _ Session = &DB{}

// Session 代表一个抽象的概念，即会话
// 暂时做成私有的，后面考虑重构，因为这个东西用户可能有点难以理解
type Session interface {
	getCore() core
	queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	execContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type Tx struct {
	tx *sql.Tx
	db *DB
	// 事务扩散方案里面，
	// 这个要在 commit 或者 rollback 的时候修改为 true
	// done bool
}

func (t *Tx) getCore() core {
	return t.db.core
}

func (t *Tx) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

func (t *Tx) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

func (t *Tx) Commit() error {
	return t.tx.Commit()
}

func (t *Tx) Rollback() error {
	return t.tx.Rollback()
}

func (t *Tx) RollbackIfNotCommit() error {
	err := t.tx.Rollback()
	if err != sql.ErrTxDone {
		return err
	}
	return nil
}
