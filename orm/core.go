package orm

import (
	"context"
	"database/sql"
	"github.com/coderi421/kyuu/orm/internal/valuer"
	"github.com/coderi421/kyuu/orm/model"
)

type core struct {
	dialect    Dialect
	r          model.Registry // 存储数据库表和 struct 映射关系的实例
	valCreator valuer.Creator // 与DB交互映射的实现
	mdls       []Middleware
}

func getHandler[T any](ctx context.Context,
	sess session,
	c core,
	qc *QueryContext) *QueryResult {
	q, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}

	// s.db 是我们定义的 DB
	// s.db.db 则是 sql.DB
	// 使用 QueryContext，从而和 GetMulti 能够复用处理结果集的代码
	rows, err := sess.queryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}

	if !rows.Next() {
		return &QueryResult{
			Err: ErrNoRows,
		}
	}

	// 创建与 db table 对应的 *struct
	tp := new(T)
	meta, err := c.r.Get(tp)
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}
	// 开始进行映射 db table 和 struct 的关系
	val := c.valCreator(tp, meta)
	// 使用存在映射关系的实体 val， 将 rows 中的数据 映射到 *struct[T] 中
	err = val.SetColumns(rows)
	return &QueryResult{
		Result: tp,
		Err:    err,
	}
}

func get[T any](ctx context.Context, c core, sess session, qc *QueryContext) *QueryResult {

	var handler Handler = func(ctx context.Context, qc *QueryContext) *QueryResult {
		// 获取跟节点
		return getHandler[T](ctx, sess, c, qc)
	}
	ms := c.mdls
	for i := len(ms) - 1; i >= 0; i-- {
		handler = ms[i](handler)
	}
	return handler(ctx, qc)
}

func exec(ctx context.Context, sess session, c core, qc *QueryContext) Result {

	var handler Handler = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return execHandler(ctx, sess, qc)
	}

	ms := c.mdls
	for i := len(ms) - 1; i >= 0; i-- {
		handler = ms[i](handler)
	}
	qr := handler(ctx, qc)
	var res sql.Result
	if qr.Result != nil {
		res = qr.Result.(sql.Result)
	}
	return Result{err: qr.Err, res: res}
}

func execHandler(ctx context.Context, sess session, qc *QueryContext) *QueryResult {
	q, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}

	res, err := sess.execContext(ctx, q.SQL, q.Args...)

	return &QueryResult{
		Result: res,
		Err:    err,
	}
}
