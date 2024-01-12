package o11y

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/util/annotations"
)

var runOnceQuerierO11y sync.Once

var querierLatency *prometheus.HistogramVec
var querierTotal *prometheus.CounterVec

type QuerierO11y struct {
	name          string
	typeOfQuerier string
	metricz       *prometheus.Registry
	wrapped       storage.Querier
}

func NewQuerierO11y(metricz *prometheus.Registry, name string, typeOfQuerier string,
	wrapped storage.Querier) *QuerierO11y {

	registerMetrics(metricz)

	return &QuerierO11y{
		name:          name,
		typeOfQuerier: typeOfQuerier,
		metricz:       metricz,
		wrapped:       wrapped,
	}
}

// Querier
func (qO11y *QuerierO11y) Select(ctx context.Context, sortSeries bool, hints *storage.SelectHints,
	matchers ...*labels.Matcher) storage.SeriesSet {

	start := time.Now()

	qO11y.countUpQueryTotal()
	result := qO11y.wrapped.Select(ctx, sortSeries, hints, matchers...)
	qO11y.observeQueryLatency(float64(time.Since(start).Seconds()))

	return result
}

// LabelQuerier
func (qO11y *QuerierO11y) Close() error {
	return qO11y.wrapped.Close()
}

// LabelQuerier
func (qO11y *QuerierO11y) LabelValues(ctx context.Context, name string,
	matchers ...*labels.Matcher) ([]string, annotations.Annotations, error) {
	return qO11y.wrapped.LabelValues(ctx, name, matchers...)
}

// LabelQuerier
func (qO11y *QuerierO11y) LabelNames(ctx context.Context,
	matchers ...*labels.Matcher) ([]string, annotations.Annotations, error) {
	return qO11y.wrapped.LabelNames(ctx, matchers...)
}

func registerMetrics(metricz *prometheus.Registry) {
	runOnceQuerierO11y.Do(func() {

		querierLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "graviola",
			Subsystem: "querier",
			Name:      "query_latency_seconds",
			Help:      "Latency of outgoing requests to a remote/group, in seconds. Only PromQL queries. Label queries are not accounted here.",
			Buckets:   []float64{0.1, 0.25, 0.5, 0.75, 1.0, 1.5, 2.5, 5.0, 10.0, 20.0, 30.0, 45.0, 60.0},
		},
			[]string{"querier_type", "querier_name"})

		querierTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "graviola",
			Subsystem: "querier",
			Name:      "query_total",
			Help:      "Counter for outgoing requests. Label queries are not accounted here.",
		},
			[]string{"querier_type", "querier_name"})

		metricz.MustRegister(querierLatency, querierTotal)
	})
}

func (qO11y *QuerierO11y) observeQueryLatency(value float64) {
	querierLatency.WithLabelValues(qO11y.typeOfQuerier, qO11y.name).Observe(value)
}

func (qO11y *QuerierO11y) countUpQueryTotal() {
	querierTotal.WithLabelValues(qO11y.typeOfQuerier, qO11y.name).Inc()
}
