package accesslog

import (
	"encoding/json"
	"github.com/coderi421/kyuu"
)

type MiddlewareBuilder struct {
	logFunc func(log string)
}

func NewBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{}
}

// LogFunc 这里如果需要配置的参数比较多，可以使用 函数选项模式
func (m *MiddlewareBuilder) LogFunc(fn func(log string)) *MiddlewareBuilder {
	m.logFunc = fn
	return m
}

func (m *MiddlewareBuilder) Build() kyuu.Middleware {
	return func(next kyuu.HandleFunc) kyuu.HandleFunc {
		return func(ctx *kyuu.Context) {
			// 要记录的请求
			defer func() {
				l := accessLog{
					Host:       ctx.Req.Host,
					Route:      ctx.MatchedRoute,
					HTTPMethod: ctx.Req.Method,
					Path:       ctx.Req.URL.Path,
				}

				data, _ := json.Marshal(l)

				m.logFunc(string(data))
			}()

			// 下一步要执行的逻辑
			next(ctx)
		}
	}
}

type accessLog struct {
	Host       string `json:"host,omitempty"`
	Route      string `json:"route,omitempty"` // 命中路由
	HTTPMethod string `json:"http_method,omitempty"`
	Path       string `json:"path,omitempty"` // 访问的路径
}
