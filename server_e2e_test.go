//go:build e2e

package kyuu

import (
	"fmt"
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	h := &HTTPServer{}

	h.AddRoute(http.MethodGet, "/user", func(ctx *Context) {
	})

	h.Get("/user", func(ctx *Context) {
		fmt.Println("处理第一件事")
	})

	_ = h.Start(":8081")
}
