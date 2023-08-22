//go:build e2e

package prometheus

import (
	"github.com/coderi421/kyuu"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

func TestMiddlewareBuilder_Build(t *testing.T) {
	builder := MiddlewareBuilder{
		Namespace: "geekbang",
		Subsystem: "web",
		Name:      "http_response",
	}
	server := kyuu.NewHTTPServer()

	server.Use(builder.Build())
	server.Get("/user", func(ctx *kyuu.Context) {
		val := rand.Intn(1000) + 1
		time.Sleep(time.Duration(val) * time.Millisecond)
		ctx.RespJSON(200, User{Name: "Tom"})
	})

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8082", nil)
	}()

	server.Start(":8081")
}

type User struct {
	Name string
}
