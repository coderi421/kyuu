package orm

import (
	"context"
	"database/sql"
)

type RawQuerier[T any] struct {
	core
	sess Session
	sql  string
	args []any
}

func RawQuery[T any](sess Session, query string, args ...any) *RawQuerier[T] {
	c := sess.getCore()
	return &RawQuerier[T]{
		core: c,
		sess: sess,
		sql:  query,
		args: args,
	}
}

func (r *RawQuerier[T]) Build() (*Query, error) {
	return &Query{
		SQL:  r.sql,
		Args: r.args,
	}, nil
}

func (r *RawQuerier[T]) Exec(ctx context.Context) Result {
	var err error
	r.model, err = r.r.Get(new(T))
	if err != nil {
		return Result{
			err: err,
		}
	}

	res := exec(ctx, r.sess, r.core, &QueryContext{
		Type:    "RAW",
		Builder: r,
		Model:   r.model,
	})

	// var t *T
	// if val, ok := res.Result.(*T); ok {
	// 	t = val
	// }
	// return t, res.Err

	var sqlRes sql.Result
	if res.Result != nil {
		sqlRes = res.Result.(sql.Result)
	}
	return Result{
		err: res.Err,
		res: sqlRes,
	}
}

// Get RawQuery 创建一个 RawQuerier 实例
// 泛型参数 T 是目标类型。
// 例如，如果查询 User 的数据，那么 T 就是 User
func (r *RawQuerier[T]) Get(ctx context.Context) (*T, error) {
	var err error
	// 获取 model 在中间件中使用
	r.model, err = r.r.Get(new(T))
	if err != nil {
		return nil, err
	}

	res := get[T](ctx, r.core, r.sess, &QueryContext{
		Type:    "RAW",
		Builder: r,
		Model:   r.model,
	})

	if res.Result != nil {
		return res.Result.(*T), err
	}
	return nil, res.Err
}

func (r *RawQuerier[T]) GetMulti(ctx context.Context) ([]*T, error) {
	//TODO implement me
	panic("implement me")
}
