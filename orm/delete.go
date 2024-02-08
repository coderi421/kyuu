package orm

import (
	"context"
	"database/sql"
)

type Deleter[T any] struct {
	builder

	table string
	where []Predicate
	core
	//	db      *DB      // 注册映射关系的实例，以及使用哪种映射方法的实例，以及 DB 实例
	sess session // db is the DB instance used for executing the
}

// NewSelector creates a new instance of Selector.
func NewDeleter[T any](sess session) *Deleter[T] {
	c := sess.getCore()
	return &Deleter[T]{
		core: c,
		sess: sess,
		builder: builder{
			dialect: c.dialect,
			quoter:  c.dialect.quoter(),
		},
	}
}

// Build generates a DELETE query based on the provided parameters.
// It returns the generated query string and any associated arguments,
// or an error if there was a problem building the query.
func (d *Deleter[T]) Build() (*Query, error) {
	var (
		t   T
		err error
	)

	// 从缓存中读取model
	d.model, err = d.r.Get(&t)
	if err != nil {
		return nil, err
	}

	_, _ = d.sb.WriteString("DELETE FROM ")

	// If the table name is not provided, use the name of the T struct.
	if d.table == "" {
		d.quote(d.model.TableName)
	} else {
		d.sb.WriteString(d.table)
	}

	// If there are any WHERE clauses, add them to the query.
	if len(d.where) > 0 {
		d.sb.WriteString(" WHERE ")
		err = d.buildPredicates(d.where)
		if err != nil {
			return nil, err
		}
	}

	d.sb.WriteByte(';')
	return &Query{
		SQL:  d.sb.String(),
		Args: d.args,
	}, nil
}

// From sets the table for the Deleter and returns a pointer to the Deleter.
// The table parameter specifies the name of the table to delete from.
func (d *Deleter[T]) From(table string) *Deleter[T] {
	d.table = table
	return d
}

// Where accepts predicates and adds them to the Deleter's where clause.
//
// Parameters:
// predicates: A list of predicates to add to the where clause.
//
// Returns:
// *Deleter[T]: The Deleter object with the updated where clause.
func (d *Deleter[T]) Where(predicates ...Predicate) *Deleter[T] {
	d.where = predicates
	return d
}

func (d *Deleter[T]) Exec(ctx context.Context) Result {
	var handler Handler = d.execHandler
	middlewares := d.mdls
	for j := len(middlewares) - 1; j >= 0; j-- {
		handler = middlewares[j](handler)
	}

	qc := &QueryContext{
		Builder: d,
		Type:    "DELETE",
	}

	res := handler(ctx, qc)
	if res.Result != nil {
		return Result{
			err: res.Err,
			res: res.Result.(sql.Result),
		}
	}

	return Result{
		err: res.Err,
	}
}

func (d *Deleter[T]) execHandler(ctx context.Context, qc *QueryContext) *QueryResult {
	q, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}

	res, err := d.sess.execContext(ctx, q.SQL, q.Args...)

	return &QueryResult{
		Result: res,
		Err:    err,
	}
}
