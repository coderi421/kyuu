package opentelemetry

import (
	"github.com/coderi421/kyuu"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/coderi421/kyuu/meddleware/opentelemetry"

type MiddlewareBuilder struct {
	Tracer trace.Tracer
}

// 可以由用户传递进来
// func NewMiddlewareBuilder(tracer trace.Tracer) *MiddlewareBuilder {
// 	return &MiddlewareBuilder{
// 		Tracer: tracer,
// 	}
// }

func (m *MiddlewareBuilder) Build() kyuu.Middleware {
	if m.Tracer == nil {
		// 创建 tracer 实例
		m.Tracer = otel.GetTracerProvider().Tracer(instrumentationName)
	}
	return func(next kyuu.HandleFunc) kyuu.HandleFunc {
		return func(ctx *kyuu.Context) {
			reqCtx := ctx.Req.Context()
			// 尝试 将客户端的 trace 结合在一起
			reqCtx = otel.GetTextMapPropagator().Extract(reqCtx, propagation.HeaderCarrier(ctx.Req.Header))

			reqCtx, span := m.Tracer.Start(reqCtx, "unknown")
			defer span.End()

			// defer func() {
			// 	// 这个是只有执行完 next 才可能有值
			// 	span.SetName(ctx.MatchedRoute)
			//
			// 	// 把响应码加上去
			// 	span.SetAttributes(attribute.Int("http.status", ctx.RespStatusCode))
			// 	span.End()
			// }()

			span.SetAttributes(attribute.String("http.method", ctx.Req.Method))
			span.SetAttributes(attribute.String("http.url", ctx.Req.URL.String()))
			span.SetAttributes(attribute.String("http.scheme", ctx.Req.URL.Scheme))
			span.SetAttributes(attribute.String("http.host", ctx.Req.Host))

			// 再次封装 req
			ctx.Req = ctx.Req.WithContext(reqCtx)

			// 直接调用下一步
			next(ctx)

			// 重命名 span 的内容 unknown -> ctx.MatchedRoute
			span.SetName(ctx.MatchedRoute)

			// 把响应码加上去
			span.SetAttributes(attribute.Int("http.status", ctx.RespStatusCode))
		}
	}
}
