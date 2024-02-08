package querylog

import (
	"context"
	"github.com/coderi421/kyuu/orm"
	"log"
)

type MiddlewareBuilder struct {
	logFunc func(query string, args []any)
	// logFunc func(query string, args...)
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{
		logFunc: func(query string, args []any) {
			log.Printf("sql: %s, args: %v", query, args)
		},
	}
}

func (m *MiddlewareBuilder) LogFunc(fn func(query string, args []any)) *MiddlewareBuilder {
	m.logFunc = fn
	return m
}

func (m MiddlewareBuilder) Build() orm.Middleware {
	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			q, err := qc.Builder.Build()
			if err != nil {
				// 要考虑记录下来吗？
				return &orm.QueryResult{
					Err: err,
				}
			}
			m.logFunc(q.SQL, q.Args)
			// 不调用 next 就是 dry run
			res := next(ctx, qc)
			return res
		}
	}
}
