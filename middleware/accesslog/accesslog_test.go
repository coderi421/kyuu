package accesslog

import (
	"fmt"
	"github.com/coderi421/kyuu"
	"net/http"
	"testing"
)

func TestMiddlewareBuilder(t *testing.T) {
	builder := MiddlewareBuilder{}
	mdl := builder.LogFunc(func(log string) {
		fmt.Println(log)
	}).Build()

	server := kyuu.NewHTTPServer()
	server.Use(mdl)
	server.Post("/a/b/*", func(ctx *kyuu.Context) {
		fmt.Println("用户逻辑")
	})
	req, err := http.NewRequest(http.MethodPost, "/a/b/c", nil)
	if err != nil {
		t.Fatal(err)
	}
	server.ServeHTTP(nil, req)
}
