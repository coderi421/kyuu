package querylog

import (
	"context"
	"errors"
	"github.com/coderi421/kyuu/orm"
	"strings"
)

// 要强制查询语句
// 1. SELECT、update、delete 必须要带 WHERE
// 2. update 和 delete 必须要带 WHERE
type MiddlewareBuilder struct {
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{}
}

func (m MiddlewareBuilder) Build() orm.Middleware {
	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			if qc.Type == "SELECT" || qc.Type == "INSERT" {
				return next(ctx, qc)
			}
			q, err := qc.Builder.Build()
			if err != nil {
				return &orm.QueryResult{
					Err: err,
				}
			}
			if strings.Contains(q.SQL, "WHERE") {
				return &orm.QueryResult{
					Err: errors.New("不准执行没有 WHERE 的 delete 或者 update 语句"),
				}
			}
			return next(ctx, qc)
		}
	}
}
