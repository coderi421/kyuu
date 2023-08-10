//go:build e2e

package kyuu

import (
	"testing"
)

func TestServer(t *testing.T) {
	s := NewHTTPServer()

	s.Get("/", func(ctx *Context) {
		ctx.Resp.Write([]byte("hello, world"))
	})
	s.Get("/user", func(ctx *Context) {
		ctx.Resp.Write([]byte("hello, user"))
	})

	_ = s.Start(":8081")
}
