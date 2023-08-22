package prometheus

import (
	"github.com/coderi421/kyuu"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
)

type MiddlewareBuilder struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
}

func (m MiddlewareBuilder) Build() kyuu.Middleware {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:      m.Name,
		Subsystem: m.Subsystem,
		Namespace: m.Name,
		Help:      m.Help,
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.90:  0.01,
			0.99:  0.001,  // 99 线
			0.999: 0.0001, // 999 线
		},
	}, []string{"pattern", "method", "status"})

	prometheus.MustRegister(vector)

	return func(next kyuu.HandleFunc) kyuu.HandleFunc {
		return func(ctx *kyuu.Context) {
			// 开始时间
			startTime := time.Now()
			// defer 算结束时间
			defer func() {
				duration := time.Now().Sub(startTime).Milliseconds()
				pattern := ctx.MatchedRoute
				if pattern == "" {
					pattern = "unknown"
				}
				vector.WithLabelValues(pattern, ctx.Req.Method,
					strconv.Itoa(ctx.RespStatusCode)).Observe(float64(duration))
			}()
			next(ctx)
		}
	}
}
