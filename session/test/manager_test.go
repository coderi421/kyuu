package test

import (
	"github.com/coderi421/kyuu"
	"github.com/coderi421/kyuu/session"
	"github.com/coderi421/kyuu/session/cookie"
	"github.com/coderi421/kyuu/session/memory"
	"github.com/google/uuid"
	"net/http"
	"testing"
	"time"
)

func TestManager(t *testing.T) {
	s := kyuu.NewHTTPServer()

	m := session.Manager{
		Store: memory.NewStore(30 * time.Minute),
		Propagator: cookie.NewPropagator("sessid", cookie.WithCookieOption(func(c *http.Cookie) {
			c.HttpOnly = true
		})),
		SessCtxKey: "_sess",
	}

	s.Get("/login", func(ctx *kyuu.Context) {
		id := uuid.New()
		sess, err := m.InitSession(ctx, id.String())
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			return
		}
		err = sess.Set(ctx.Req.Context(), "mykey", "some value")
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			return
		}
	})

	s.Get("/resource", func(ctx *kyuu.Context) {
		sess, err := m.GetSession(ctx)
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			return
		}
		val, err := sess.Get(ctx.Req.Context(), "mykey")
		ctx.RespData = []byte(val)
	})

	s.Get("/logout", func(ctx *kyuu.Context) {
		_ = m.RemoveSession(ctx)
	})

	s.Use(func(next kyuu.HandleFunc) kyuu.HandleFunc {
		return func(ctx *kyuu.Context) {
			// 执行校验
			if ctx.Req.URL.Path != "/login" {
				sess, err := m.GetSession(ctx)
				// 不管发生了什么错误，对于用户我们都是返回未授权
				if err != nil {
					ctx.RespStatusCode = http.StatusUnauthorized
					return
				}
				ctx.UserValues[m.SessCtxKey] = sess
				_ = m.Refresh(ctx.Req.Context(), sess.ID())
			}
			next(ctx)
		}
	})

	s.Start(":8081")
}
