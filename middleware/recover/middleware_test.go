package recover

import (
	"fmt"
	"github.com/coderi421/kyuu"
	"testing"
)

func TestMiddlewareBuilder_Builder(t *testing.T) {
	builder := MiddlewareBuilder{
		StatusCode: 500,
		Data:       []byte("panic 了"),
		Log: func(ctx *kyuu.Context, err any) {
			fmt.Printf("panic 路径: %s", ctx.Req.URL.String())
		},
	}

	server := kyuu.NewHTTPServer()
	server.Use(builder.Builder())
	server.Get("/user", func(ctx *kyuu.Context) {
		panic("发生panic 了")
	})
	server.Start(":8081")
}
