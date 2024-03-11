package orm

import (
	"context"
	"github.com/coderi421/kyuu/orm/model"
)

// QueryContext 中间件的上下文，冗余了 Builder model 等，是因为还没有执行 sql 前，有的中间件，需要使用这些信息， 这里优化还可以考虑 构建 builder 后的 sql 拼接结果，省的每次调用都需要调用 builder 进行拼接 sql, 这里没这么做，可能因为怕别人篡改？
type QueryContext struct {
	// Type 声明查询类型。即 SELECT, UPDATE, DELETE 和 INSERT
	Type string

	// builder 使用的时候，大多数情况下你需要转换到具体的类型
	// 才能篡改查询
	Builder QueryBuilder
	// qc.Model.TableName 为了有的中间件在拦截时需要 Model 信息
	// 所以需要冗余一份在 middleware 的上下文中
	Model *model.Model
}

type QueryResult struct {
	// Result 在不同的查询里面，类型是不同的
	// Selector.Get 里面，这会是单个结果
	// Selector.GetMulti，这会是一个切片
	// 其它情况下，它会是 Result 类型
	Result any
	Err    error
}

type Middleware func(next Handler) Handler

type Handler func(ctx context.Context, qc *QueryContext) *QueryResult
