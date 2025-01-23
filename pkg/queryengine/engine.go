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

// This is a thin wrapper of Prometheus query engine, used to make it easier to debug and add
// telemetry
type GraviolaQueryEngine struct {
	logger             *slog.Logger
	wrappedQueryEngine *promql.Engine
}

func NewGraviolaQueryEngine(
	logger *slog.Logger, metricRegistry *prometheus.Registry, conf config.GraviolaConfig,
) *GraviolaQueryEngine {
	wrappedPromQLEngine := promql.NewEngine(promql.EngineOpts{
		Timeout:              conf.QueryConf.TimeoutDuration(),
		MaxSamples:           conf.QueryConf.MaxSamples,
		LookbackDelta:        conf.QueryConf.LookbackDeltaDuration(),
		EnableAtModifier:     true,
		EnableNegativeOffset: true,
		ActiveQueryTracker:   querytracker.NewGraviolaQueryTracker(conf.QueryConf.ConcurrentQueries),
		Reg:                  metricRegistry,
		Logger:               logger,
	})

	return &GraviolaQueryEngine{
		logger:             logger,
		wrappedQueryEngine: wrappedPromQLEngine,
	}
}

// QueryEngine
func (gravQueryEng *GraviolaQueryEngine) SetQueryLogger(_ promql.QueryLogger) {
	panic("should not be called")
}

// QueryEngine
func (gravQueryEng *GraviolaQueryEngine) NewInstantQuery(
	ctx context.Context, q storage.Queryable, opts promql.QueryOpts, qs string, ts time.Time,
) (promql.Query, error) {
	return gravQueryEng.wrappedQueryEngine.NewInstantQuery(ctx, q, opts, qs, ts)
}

// QueryEngine
func (gravQueryEng *GraviolaQueryEngine) NewRangeQuery(
	ctx context.Context, q storage.Queryable, opts promql.QueryOpts, qs string, start, end time.Time,
	interval time.Duration,
) (promql.Query, error) {
	return gravQueryEng.wrappedQueryEngine.NewRangeQuery(ctx, q, opts, qs, start, end, interval)
}
