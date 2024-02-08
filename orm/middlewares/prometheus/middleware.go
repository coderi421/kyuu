package prometheus

import (
	"context"
	"github.com/coderi421/kyuu/orm"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type MiddlewareBuilder struct {
	Namespace   string
	Subsystem   string
	Name        string
	Help        string
	ConstLabels map[string]string
}

func (m MiddlewareBuilder) Build() orm.Middleware {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:        m.Name,
		Subsystem:   m.Subsystem,
		Namespace:   m.Name,
		Help:        m.Help,
		ConstLabels: m.ConstLabels,
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.90:  0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, []string{"type", "table"})

	prometheus.MustRegister(vector)

	// errCounterVec 记录错误数
	// histogram 也可以
	// active query
	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			startTime := time.Now()
			defer func() {
				// 执行时间
				vector.WithLabelValues(qc.Type, qc.Model.TableName).Observe(float64(time.Since(startTime).Milliseconds()))
			}()
			return next(ctx, qc)
		}
	}
}
