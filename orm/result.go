package orm

import "database/sql"

type Result struct {
	err error
	res sql.Result
}

// LastInsertId 重新 database sql 的 Result 方法 做一层拦截
func (r Result) LastInsertId() (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return r.res.LastInsertId()
}

func (r Result) RowsAffected() (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return r.res.RowsAffected()
}

func (r Result) Err() error {
	return r.err
}
