package errhdl

import "github.com/coderi421/kyuu"

type MiddlewareBuilder struct {
	// 这种设计只能返回固定的值
	// 不能做到动态渲染
	resp map[int][]byte
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{
		resp: map[int][]byte{},
	}
}

// AddCode 注册响应码拦截 code
func (m *MiddlewareBuilder) AddCode(status int, data []byte) *MiddlewareBuilder {
	m.resp[status] = data
	return m
}

func (m *MiddlewareBuilder) Build() kyuu.Middleware {
	return func(next kyuu.HandleFunc) kyuu.HandleFunc {
		return func(ctx *kyuu.Context) {
			next(ctx)
			resp, ok := m.resp[ctx.RespStatusCode]
			if ok {
				// 值修改 RespData, 这样其他中间件还能继续操作
				ctx.RespData = resp
			}
		}
	}
}
