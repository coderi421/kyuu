package recover

import "github.com/coderi421/kyuu"

type MiddlewareBuilder struct {
	StatusCode int
	Data       []byte
	LogFunc    func(ctx *kyuu.Context, err any)
	// log func(err any)
	// LogFunc func(ctx *web.Context)
	// log func(stack string)
}

func (m *MiddlewareBuilder) Builder() kyuu.Middleware {
	return func(next kyuu.HandleFunc) kyuu.HandleFunc {
		return func(ctx *kyuu.Context) {
			defer func() {
				if err := recover(); err != nil {
					ctx.RespData = m.Data
					ctx.RespStatusCode = m.StatusCode
					// 万一 LogFunc 也panic，那我们也无能为力了
					m.LogFunc(ctx, err)
				}
			}()
			next(ctx)
		}
	}
}
