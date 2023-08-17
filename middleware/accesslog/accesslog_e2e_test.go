package accesslog

import (
	"fmt"
	"github.com/coderi421/kyuu"
	"testing"
)

func TestMiddlewareBuilderE2E(t *testing.T) {
	builder := MiddlewareBuilder{}
	mdl := builder.LogFunc(func(log string) {
		fmt.Println("it`s access log test")
		fmt.Println(log)
	}).Build()

	server := kyuu.NewHTTPServer()
	server.Use(mdl)
	server.Get("/a/b/*", func(ctx *kyuu.Context) {
		ctx.Resp.Write([]byte("hello"))
	})

	err := server.Start(":8081")
	if err != nil {
		return
	}
}
