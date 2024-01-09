package queryengine_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/jademcosta/graviola/pkg/graviolalog"
	"github.com/jademcosta/graviola/pkg/queryengine"
	"github.com/jademcosta/graviola/pkg/storageproxy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/storage"
	"github.com/stretchr/testify/assert"
)

func TestSampleLimit(t *testing.T) {
	logger := graviolalog.NewLogger(conf.LogConf)
	ctx := context.Background()
	currentTime := time.Now()

	testCases := []struct {
		caseName    string
		returnSet   storage.SeriesSet
		maxSamples  int
		shouldError bool
	}{
		{"is empty",
			storage.NoopSeriesSet(),
			10,
			false,
		},
		{"is less than the limit",
			&domain.GraviolaSeriesSet{
				Series: []*domain.GraviolaSeries{
					{
						Lbs: labels.FromStrings("__name__", "up"),
						Datapoints: []model.SamplePair{
							{Timestamp: model.Time(currentTime.UnixMilli()), Value: 1.1},
						},
					},
				},
			},
			2,
			false,
		},
		{"is less than the limit (datapoints doesn't count, only time-series)",
			&domain.GraviolaSeriesSet{
				Series: []*domain.GraviolaSeries{
					{
						Lbs: labels.FromStrings("__name__", "up"),
						Datapoints: []model.SamplePair{
							{Timestamp: model.Time(currentTime.Add(-90 * time.Second).UnixMilli()), Value: 1.1},
							{Timestamp: model.Time(currentTime.Add(-60 * time.Second).UnixMilli()), Value: 1.2},
							{Timestamp: model.Time(currentTime.Add(-30 * time.Second).UnixMilli()), Value: 1.3},
							{Timestamp: model.Time(currentTime.UnixMilli()), Value: 1.4},
						},
					},
				},
			},
			2,
			false,
		},
		{"is more than the limit",
			&domain.GraviolaSeriesSet{
				Series: []*domain.GraviolaSeries{
					{
						Lbs: labels.FromStrings("__name__", "up"),
						Datapoints: []model.SamplePair{
							{Timestamp: model.Time(currentTime.UnixMilli()), Value: 1.1},
						},
					},
					{
						Lbs: labels.FromStrings("__name__", "up", "instance", "127.0.0.1"),
						Datapoints: []model.SamplePair{
							{Timestamp: model.Time(currentTime.UnixMilli()), Value: 1.4},
						},
					},
					{
						Lbs: labels.FromStrings("__name__", "up", "instance", "192.168.0.0"),
						Datapoints: []model.SamplePair{
							{Timestamp: model.Time(currentTime.UnixMilli()), Value: 1.5},
						},
					},
				},
			},
			2,
			true,
		},
		{"is more than the limit",
			&domain.GraviolaSeriesSet{
				Series: []*domain.GraviolaSeries{
					{
						Lbs: labels.FromStrings("__name__", "up"),
						Datapoints: []model.SamplePair{
							{Timestamp: model.Time(currentTime.UnixMilli()), Value: 1.1},
						},
					},
					{
						Lbs: labels.FromStrings("__name__", "up", "instance", "127.0.0.1"),
						Datapoints: []model.SamplePair{
							{Timestamp: model.Time(currentTime.UnixMilli()), Value: 1.4},
						},
					},
					{
						Lbs: labels.FromStrings("__name__", "up", "instance", "192.168.0.0"),
						Datapoints: []model.SamplePair{
							{Timestamp: model.Time(currentTime.UnixMilli()), Value: 1.5},
						},
					},
					{
						Lbs: labels.FromStrings("__name__", "up", "instance", "192.168.0.1"),
						Datapoints: []model.SamplePair{
							{Timestamp: model.Time(currentTime.UnixMilli()), Value: 1.6},
						},
					},
					{
						Lbs: labels.FromStrings("__name__", "up", "instance", "192.168.0.2"),
						Datapoints: []model.SamplePair{
							{Timestamp: model.Time(currentTime.UnixMilli()), Value: 1.7},
						},
					},
				},
			},
			2,
			true,
		},
	}

	for _, tc := range testCases {
		reg := prometheus.NewRegistry()

		conf := config.GraviolaConfig{
			ApiConf: config.ApiConfig{
				Timeout: "3m",
			},
			QueryConf: config.QueryConfig{
				MaxSamples:        tc.maxSamples,
				LookbackDelta:     config.DefaultQueryLookbackDelta,
				ConcurrentQueries: 2,
			},
		}

		mock1 := &MockQuerier{
			selectReturn: tc.returnSet,
		}

		gravStorage := storageproxy.NewGraviolaStorage(logger, []storage.Querier{mock1})
		eng := queryengine.NewGraviolaQueryEngine(logger, reg, conf)

		querier, err := eng.NewInstantQuery(ctx, gravStorage, promql.NewPrometheusQueryOpts(false, 0), "up", currentTime)
		assert.NoError(t, err, "should return no error")

		result := querier.Exec(ctx)
		if tc.shouldError {
			assert.Error(t, result.Err, "query should error when returned series %s", tc.caseName)
			assert.ErrorIs(t, promql.ErrTooManySamples("query execution"), result.Err, "should be a specific type of error")
		} else {
			assert.NoError(t, result.Err, "query should NOT error when returned series %s", tc.caseName)
		}
	}
}

func TestEngineUsesTheProvidedLookbackDelta(t *testing.T) {
	logger := graviolalog.NewLogger(conf.LogConf)
	reg := prometheus.NewRegistry()
	ctx := context.Background()
	currentTime := time.Now()

	conf := config.GraviolaConfig{
		ApiConf: config.ApiConfig{
			Timeout: "3m",
		},
		QueryConf: config.QueryConfig{
			MaxSamples:        10,
			LookbackDelta:     "17s",
			ConcurrentQueries: 2,
		},
	}

	mock1 := &MockQuerier{
		selectReturn: storage.NoopSeriesSet(),
	}

	gravStorage := storageproxy.NewGraviolaStorage(logger, []storage.Querier{mock1})
	eng := queryengine.NewGraviolaQueryEngine(logger, reg, conf)

	querier, err := eng.NewInstantQuery(ctx, gravStorage, promql.NewPrometheusQueryOpts(false, 0), "up", currentTime)
	assert.NoError(t, err, "should return no error")

	querier.Exec(ctx)

	assert.Equal(
		t,
		selectCalled{sortSeries: false, hints: &storage.SelectHints{
			End:   currentTime.UnixMilli(),
			Start: currentTime.Add(-17 * time.Second).UnixMilli(),
		}, matchers: []*labels.Matcher{labels.MustNewMatcher(labels.MatchEqual, "__name__", "up")}},
		mock1.selectCalledWith[0],
	)
}

func TestEngineConcurrentQueriesLimit(t *testing.T) {
	logger := graviolalog.NewLogger(conf.LogConf)
	reg := prometheus.NewRegistry()
	ctx := context.Background()
	currentTime := time.Now()

	conf := config.GraviolaConfig{
		ApiConf: config.ApiConfig{
			Timeout: "3m",
		},
		QueryConf: config.QueryConfig{
			MaxSamples:        10,
			LookbackDelta:     "17s",
			ConcurrentQueries: 1,
		},
	}

	var wg sync.WaitGroup
	wg.Add(2)

	mock1 := &MockQuerier{
		selectReturn: storage.NoopSeriesSet(),
		delay:        200 * time.Millisecond,
	}

	gravStorage := storageproxy.NewGraviolaStorage(logger, []storage.Querier{mock1})
	eng := queryengine.NewGraviolaQueryEngine(logger, reg, conf)

	querier, err := eng.NewInstantQuery(ctx, gravStorage, promql.NewPrometheusQueryOpts(false, 0), "up", currentTime)
	assert.NoError(t, err, "should return no error")

	start := time.Now()

	querier.Exec(ctx)
	wg.Done()

	go func() {
		querier, err := eng.NewInstantQuery(ctx, gravStorage, promql.NewPrometheusQueryOpts(false, 0), "up", currentTime)
		assert.NoError(t, err, "should return no error")
		querier.Exec(ctx)
		wg.Done()
	}()

	wg.Wait()

	elapsed := time.Since(start)

	assert.GreaterOrEqual(t, elapsed, 400*time.Millisecond, "should have respected the concurrent queries limit")
}
