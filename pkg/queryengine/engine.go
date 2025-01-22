package queryengine

import (
	"context"
	"log/slog"
	"time"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/querytracker"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/storage"
)

// This is a thin wrapper, used to make it easier to debug
type GraviolaQueryEngine struct {
	logger  *slog.Logger
	wrapped *promql.Engine
}

func NewGraviolaQueryEngine(logger *slog.Logger, metricRegistry *prometheus.Registry, conf config.GraviolaConfig) *GraviolaQueryEngine {
	wrapped := promql.NewEngine(promql.EngineOpts{
		Timeout:              conf.APIConf.TimeoutDuration(),
		MaxSamples:           conf.QueryConf.MaxSamples,
		LookbackDelta:        conf.QueryConf.LookbackDeltaDuration(),
		EnableAtModifier:     true,
		EnableNegativeOffset: true,
		ActiveQueryTracker:   querytracker.NewGraviolaQueryTracker(conf.QueryConf.ConcurrentQueries),
		Reg:                  metricRegistry,
		Logger:               logger,
	})

	return &GraviolaQueryEngine{
		logger:  logger,
		wrapped: wrapped,
	}
}

// QueryEngine
func (gravEng *GraviolaQueryEngine) SetQueryLogger(_ promql.QueryLogger) {
	panic("should not be called")
}

// QueryEngine
func (gravEng *GraviolaQueryEngine) NewInstantQuery(
	ctx context.Context, q storage.Queryable, opts promql.QueryOpts, qs string, ts time.Time,
) (promql.Query, error) {
	return gravEng.wrapped.NewInstantQuery(ctx, q, opts, qs, ts)
}

// QueryEngine
func (gravEng *GraviolaQueryEngine) NewRangeQuery(
	ctx context.Context, q storage.Queryable, opts promql.QueryOpts, qs string, start, end time.Time, interval time.Duration,
) (promql.Query, error) {
	return gravEng.wrapped.NewRangeQuery(ctx, q, opts, qs, start, end, interval)
}
