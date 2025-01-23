package queryengine_test

import (
	"context"
	"fmt"
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
	"github.com/prometheus/prometheus/util/annotations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var conf config.GraviolaConfig = config.GraviolaConfig{
	QueryConf: config.QueryConfig{
		MaxSamples:        1000,
		LookbackDelta:     config.DefaultQueryLookbackDelta,
		ConcurrentQueries: 1,
		Timeout:           "3m",
	},
}

var queryTestCases = []struct {
	//This is a function type only to be able to do some nil checking for cases where we have
	// no expected hints
	expectedHintsFunc     func() *storage.SelectHints
	query                 string
	expectedLabelMatchers [][]labels.Matcher
}{
	{nil, "up", [][]labels.Matcher{{{Type: labels.MatchEqual, Name: "__name__", Value: "up"}}}},
	{nil, "count(up)", [][]labels.Matcher{{{Type: labels.MatchEqual, Name: "__name__", Value: "up"}}}},
	{nil, "sum(up)", [][]labels.Matcher{{{Type: labels.MatchEqual, Name: "__name__", Value: "up"}}}},
	{nil, "rate(up[5m])", [][]labels.Matcher{{{Type: labels.MatchEqual, Name: "__name__", Value: "up"}}}},
	{nil, "sum(rate(up[5m])) by(instance)", [][]labels.Matcher{{{Type: labels.MatchEqual, Name: "__name__", Value: "up"}}}},
	{nil, `topk(4, sum(up) by(account))`, [][]labels.Matcher{{{Type: labels.MatchEqual, Name: "__name__", Value: "up"}}}},
	{
		func() *storage.SelectHints {
			return &storage.SelectHints{
				Start: time.Unix(12345678, 0).Add(-5 * time.Minute).Add(time.Millisecond).UnixMilli(),
				End:   time.Unix(12345678, 0).UnixMilli(),
			}
		},
		`sum(up @ 12345678)`, [][]labels.Matcher{{{Type: labels.MatchEqual, Name: "__name__", Value: "up"}}},
	},
	{
		func() *storage.SelectHints {
			return &storage.SelectHints{
				Start: time.Now().Add(-2 * time.Hour).Add(-5 * time.Minute).Add(time.Millisecond).UnixMilli(),
				End:   time.Now().UnixMilli(),
				Step:  300000,
			}
		},
		`sum(up) [2h:5m]`, [][]labels.Matcher{{{Type: labels.MatchEqual, Name: "__name__", Value: "up"}}},
	},
	{nil, `topk(4, sum(up{region="abc"}) by(account))`,
		[][]labels.Matcher{{
			{Type: labels.MatchEqual, Name: "region", Value: "abc"},
			{Type: labels.MatchEqual, Name: "__name__", Value: "up"},
		}}},
	{nil, "sum(rate(some_sum[5m])) by(instance) / sum(rate(some_count[5m])) by(instance)",
		[][]labels.Matcher{
			{{Type: labels.MatchEqual, Name: "__name__", Value: "some_sum"}},
			{{Type: labels.MatchEqual, Name: "__name__", Value: "some_count"}},
		},
	},
	{nil, `histogram_quantile(0.9, sum(rate(http_requests_total[5m])))`,
		[][]labels.Matcher{
			{{Type: labels.MatchEqual, Name: "__name__", Value: "http_requests_total"}},
		},
	},
	{nil, `histogram_quantile(0.9, sum(rate(http_requests_total{instance="127.0.0.1"}[5m])))`,
		[][]labels.Matcher{
			{
				{Type: labels.MatchEqual, Name: "instance", Value: "127.0.0.1"},
				{Type: labels.MatchEqual, Name: "__name__", Value: "http_requests_total"},
			},
		},
	},
	{nil, `sum(rate(some_sum[5m])) by(instance) / on(somerandomaaa) sum(rate(some_count{instance="1234"}[5m]))`,
		[][]labels.Matcher{
			{{Type: labels.MatchEqual, Name: "__name__", Value: "some_sum"}},
			{
				{Type: labels.MatchEqual, Name: "instance", Value: "1234"},
				{Type: labels.MatchEqual, Name: "__name__", Value: "some_count"},
			},
		},
	},
	{nil, `some_sum{instance="1234"} / some_count{instance="1234"}`,
		[][]labels.Matcher{
			{
				{Type: labels.MatchEqual, Name: "instance", Value: "1234"},
				{Type: labels.MatchEqual, Name: "__name__", Value: "some_sum"},
			},
			{
				{Type: labels.MatchEqual, Name: "instance", Value: "1234"},
				{Type: labels.MatchEqual, Name: "__name__", Value: "some_count"},
			},
		},
	},
	{nil, `some_count{instance="5678"} / some_count{instance="1234"}`,
		[][]labels.Matcher{
			{
				{Type: labels.MatchEqual, Name: "instance", Value: "5678"},
				{Type: labels.MatchEqual, Name: "__name__", Value: "some_count"},
			},
			{
				{Type: labels.MatchEqual, Name: "instance", Value: "1234"},
				{Type: labels.MatchEqual, Name: "__name__", Value: "some_count"},
			},
		},
	},
	{nil, `some_count{instance="5678"} + some_count{instance="1234"} * 2`,
		[][]labels.Matcher{
			{
				{Type: labels.MatchEqual, Name: "instance", Value: "5678"},
				{Type: labels.MatchEqual, Name: "__name__", Value: "some_count"},
			},
			{
				{Type: labels.MatchEqual, Name: "instance", Value: "1234"},
				{Type: labels.MatchEqual, Name: "__name__", Value: "some_count"},
			},
		},
	},
	//TODO: maybe an internal cache could help with this case
	{nil, `some_count{instance="5678"} + some_count{instance="1234"} * some_count{instance="5678"}`,
		[][]labels.Matcher{
			{
				{Type: labels.MatchEqual, Name: "instance", Value: "5678"},
				{Type: labels.MatchEqual, Name: "__name__", Value: "some_count"},
			},
			{
				{Type: labels.MatchEqual, Name: "instance", Value: "1234"},
				{Type: labels.MatchEqual, Name: "__name__", Value: "some_count"},
			},
			{
				{Type: labels.MatchEqual, Name: "instance", Value: "5678"},
				{Type: labels.MatchEqual, Name: "__name__", Value: "some_count"},
			},
		},
	},
	//TODO: maybe an internal cache could help with this case
	{nil, `some_count{instance="5678"} + some_count{instance="5678"} * some_count{instance="5678"}`,
		[][]labels.Matcher{
			{
				{Type: labels.MatchEqual, Name: "instance", Value: "5678"},
				{Type: labels.MatchEqual, Name: "__name__", Value: "some_count"},
			},
			{
				{Type: labels.MatchEqual, Name: "instance", Value: "5678"},
				{Type: labels.MatchEqual, Name: "__name__", Value: "some_count"},
			},
			{
				{Type: labels.MatchEqual, Name: "instance", Value: "5678"},
				{Type: labels.MatchEqual, Name: "__name__", Value: "some_count"},
			},
		},
	},
}

type selectCalled struct {
	sortSeries bool
	hints      *storage.SelectHints
	matchers   []*labels.Matcher
}

type MockQuerier struct {
	delay            time.Duration
	selectCalledWith []selectCalled
	selectReturn     storage.SeriesSet
}

func (mock *MockQuerier) Select(_ context.Context, sortSeries bool, hints *storage.SelectHints, matchers ...*labels.Matcher) storage.SeriesSet {
	mock.selectCalledWith = append(mock.selectCalledWith, selectCalled{
		sortSeries: sortSeries,
		hints:      hints,
		matchers:   matchers,
	})

	if mock.delay != 0 {
		time.Sleep(mock.delay)
	}
	return mock.selectReturn
}

func (mock *MockQuerier) Close() error {
	return nil
}

func (mock *MockQuerier) LabelValues(_ context.Context, _ string, _ *storage.LabelHints, _ ...*labels.Matcher) ([]string, annotations.Annotations, error) {
	return nil, nil, nil
}

func (mock *MockQuerier) LabelNames(_ context.Context, _ *storage.LabelHints, _ ...*labels.Matcher) ([]string, annotations.Annotations, error) {
	return nil, nil, nil
}

func TestIntegrationSendsExpectedHintsAndLabelMatchers(t *testing.T) {

	logger := graviolalog.NewLogger(conf.LogConf)
	ctx := context.Background()

	for _, tc := range queryTestCases {
		reg := prometheus.NewRegistry()

		mockQuerier := &MockQuerier{
			selectReturn: storage.NoopSeriesSet(),
		}

		groups := []storage.Querier{
			mockQuerier,
		}

		gravStorage := storageproxy.NewGraviolaStorage(logger, groups, defaultMergeStrategy)
		sut := queryengine.NewGraviolaQueryEngine(logger, reg, conf)

		currentTime := time.Now()

		querier, err := sut.NewInstantQuery(
			ctx, gravStorage, promql.NewPrometheusQueryOpts(false, 0), tc.query, currentTime)
		require.NoError(t, err, "should return no error")

		querier.Exec(ctx)

		queriesMatchersReceived := make([][]labels.Matcher, 0)

		for _, selectEntry := range mockQuerier.selectCalledWith {
			assert.False(t, selectEntry.sortSeries, "should not tell series should be sorted")

			if tc.expectedHintsFunc != nil {
				expectedHints := tc.expectedHintsFunc()
				assert.Equal(t, expectedHints.Step, selectEntry.hints.Step, "hints step should be equal")
				assert.InDelta(t, expectedHints.End, selectEntry.hints.End, 0.01, "hints end should be equal (inside a delta)")
				assert.InDelta(t, expectedHints.Start, selectEntry.hints.Start, 0.01, "hints start should be equal (inside a delta)")

			} else {
				assert.Equal(t, int64(0), selectEntry.hints.Step, "hints step should be zero")
				assert.Equal(t, currentTime.UnixMilli(), selectEntry.hints.End, "hints end should be equal")
				//I don't know why Prometheus adds this extra millisecond. Just add it here to make sure
				// we keep the consistency.
				assert.Equal(t, currentTime.Add(-5*time.Minute).Add(time.Millisecond).UnixMilli(),
					selectEntry.hints.Start, "hints start should be equal")
			}

			queriesMatchersReceived = append(queriesMatchersReceived, derefMatchers(selectEntry.matchers))
		}

		assert.Equalf(t, len(tc.expectedLabelMatchers), len(queriesMatchersReceived),
			"should have performed %d queries", len(tc.expectedLabelMatchers))
		assert.Equalf(t, tc.expectedLabelMatchers, queriesMatchersReceived,
			"should have performed the correct queries")
	}

	for _, tc := range queryTestCases {
		if tc.query == "sum(up) [2h:5m]" {
			continue
			//This subquery usage doesn't work on range queries
		}

		reg := prometheus.NewRegistry()

		mockQuerier := &MockQuerier{
			selectReturn: storage.NoopSeriesSet(),
		}

		groups := []storage.Querier{
			mockQuerier,
		}

		gravStorage := storageproxy.NewGraviolaStorage(logger, groups, defaultMergeStrategy)
		sut := queryengine.NewGraviolaQueryEngine(logger, reg, conf)

		currentTime := time.Now()
		startTime := currentTime.Add(-2 * time.Hour)
		endTime := currentTime
		step := time.Minute
		if tc.expectedHintsFunc != nil {
			step = time.Duration(tc.expectedHintsFunc().Step)
		}

		querier, err := sut.NewRangeQuery(
			ctx, gravStorage, promql.NewPrometheusQueryOpts(false, 0), tc.query,
			startTime, endTime, step,
		)
		if err != nil {
			panic(err)
		}

		querier.Exec(ctx)

		queriesMatchersReceived := make([][]labels.Matcher, 0)

		for _, selectEntry := range mockQuerier.selectCalledWith {
			assert.False(t, selectEntry.sortSeries, "should be equal")

			if tc.expectedHintsFunc != nil {
				localHints := tc.expectedHintsFunc()
				assert.Equal(t, localHints.Step, selectEntry.hints.Step, "hints step should be equal")
				assert.InDelta(t, localHints.End, selectEntry.hints.End, 100.0, "hints end should be equal (inside a delta)")
				assert.InDelta(t, localHints.Start, selectEntry.hints.Start, 100.0, "hints start should be equal (inside a delta)")

			} else {
				assert.Equal(t, int64(60000), selectEntry.hints.Step, "hints step should be equal")
				assert.Equal(t, endTime.UnixMilli(), selectEntry.hints.End, "hints end should be equal")
				assert.Equal(t, startTime.Add(-5*time.Minute).Add(time.Millisecond).UnixMilli(), selectEntry.hints.Start, "hints start should be equal")
			}

			queriesMatchersReceived = append(queriesMatchersReceived, derefMatchers(selectEntry.matchers))
		}

		assert.Equalf(t, len(tc.expectedLabelMatchers), len(queriesMatchersReceived), "should have performed %d queries", len(tc.expectedLabelMatchers))
		assert.Equalf(t, tc.expectedLabelMatchers, queriesMatchersReceived, "should have performed the correct queries")
	}
}

func TestIntegrationHandlesCorrectlyTheReturnedSeriesSetOnInstantQuery(t *testing.T) {

	logger := graviolalog.NewLogger(conf.LogConf)
	ctx := context.Background()

	currentTime := time.Now()

	defaultSeries := []*domain.GraviolaSeries{
		{
			Lbs: labels.FromStrings("__name__", "up", "path", "/metrics"),
			Datapoints: []model.SamplePair{
				{Timestamp: model.Time(currentTime.UnixMilli()), Value: 1.5},
			},
		},
		{
			Lbs: labels.FromStrings("__name__", "up", "path", "/healthy"),
			Datapoints: []model.SamplePair{
				{Timestamp: model.Time(currentTime.Add(-4 * time.Minute).Add(-59 * time.Second).UnixMilli()), Value: 5.0},
			},
		},
		{
			Lbs: labels.FromStrings("__name__", "up", "path", "/ready"),
			Datapoints: []model.SamplePair{
				{Timestamp: model.Time(currentTime.Add(-5 * time.Minute).Add(-1 * time.Second).UnixMilli()), Value: 555.17},
			},
		},
	}

	testCases := []struct {
		series   []*domain.GraviolaSeries
		query    string
		expected string
	}{
		{
			defaultSeries,
			`up`,
			fmt.Sprintf("{__name__=\"up\", path=\"/metrics\"} => 1.5 @[%d]\n{__name__=\"up\", path=\"/healthy\"} => 5 @[%d]",
				currentTime.UnixMilli(), currentTime.UnixMilli()),
		},
		{
			defaultSeries,
			`sum(up)`,
			fmt.Sprintf("{} => 6.5 @[%d]",
				currentTime.UnixMilli()),
		},
		{
			defaultSeries,
			`count(up)`,
			fmt.Sprintf("{} => 2 @[%d]",
				currentTime.UnixMilli()),
		},
		{
			defaultSeries,
			`sum(up) by(path)`,
			fmt.Sprintf("{path=\"/metrics\"} => 1.5 @[%d]\n{path=\"/healthy\"} => 5 @[%d]",
				currentTime.UnixMilli(), currentTime.UnixMilli()),
		},
		{
			defaultSeries,
			`sum(up{path="/healthy"}) - sum(up{path="/metrics"})`,
			"",
		},
		{ // This is correct. The filtering has to occur on the 'remote', and not on the engine
			// The engine works on the aggregations/functions level
			defaultSeries,
			`up{path="/metrics"}`,
			fmt.Sprintf("{__name__=\"up\", path=\"/metrics\"} => 1.5 @[%d]\n{__name__=\"up\", path=\"/healthy\"} => 5 @[%d]",
				currentTime.UnixMilli(), currentTime.UnixMilli()),
		},
		{
			defaultSeries,
			fmt.Sprintf(`up @ %d`, currentTime.Add(-1*time.Minute).Unix()),
			fmt.Sprintf("{__name__=\"up\", path=\"/healthy\"} => 5 @[%d]\n{__name__=\"up\", path=\"/ready\"} => 555.17 @[%d]",
				currentTime.UnixMilli(), currentTime.UnixMilli()),
		},
	}

	for _, tc := range testCases {
		reg := prometheus.NewRegistry()

		mockQuerier := &MockQuerier{
			selectReturn: &domain.GraviolaSeriesSet{
				Series: tc.series,
			},
		}

		groups := []storage.Querier{
			mockQuerier,
		}

		gravStorage := storageproxy.NewGraviolaStorage(logger, groups, defaultMergeStrategy)
		eng := queryengine.NewGraviolaQueryEngine(logger, reg, conf)

		querier, err := eng.NewInstantQuery(ctx, gravStorage, promql.NewPrometheusQueryOpts(false, 0), tc.query, currentTime)
		require.NoError(t, err, "should return no error")

		result := querier.Exec(ctx)

		assert.Equal(t, tc.expected, result.String(), "should be equal")
	}
}

func derefMatchers(matchers []*labels.Matcher) []labels.Matcher {
	derefed := make([]labels.Matcher, 0)

	for _, matcher := range matchers {
		derefed = append(derefed, *matcher)
	}
	return derefed
}

//TODO:
// Test @ symbol usage
