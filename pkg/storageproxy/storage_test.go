package storageproxy_test

import (
	"context"
	"math/rand"
	"reflect"
	"slices"
	"sync"
	"testing"

	"github.com/jademcosta/graviola/internal/mocks"
	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/jademcosta/graviola/pkg/graviolalog"
	"github.com/jademcosta/graviola/pkg/remotestoragegroup"
	"github.com/jademcosta/graviola/pkg/storageproxy"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const anyMinTime = int64(0)
const anyMaxTime = int64(1)

var logg = graviolalog.NewLogger(config.LogConfig{Level: "error"})
var defaultMergeStrategy = remotestoragegroup.MergeStrategyFactory(config.DefaultMergeStrategyType)

func TestSelect(t *testing.T) {
	mockStorage1 := &mocks.RemoteStorageMock{
		SeriesSet: &domain.GraviolaSeriesSet{
			Series: []*domain.GraviolaSeries{
				{Lbs: labels.FromStrings("label1", "val1"),
					Datapoints: []model.SamplePair{{Timestamp: 5819, Value: 5.9}}},
			},
		},
	}
	mockStorage2 := &mocks.RemoteStorageMock{
		SeriesSet: &domain.GraviolaSeriesSet{
			Series: []*domain.GraviolaSeries{
				{Lbs: labels.FromStrings("label2", "val2"),
					Datapoints: []model.SamplePair{{Timestamp: 5999, Value: 5.1}}},
			},
		},
	}

	sut := storageproxy.NewGraviolaStorage(logg, []storage.Querier{mockStorage1, mockStorage2}, defaultMergeStrategy)

	querier, err := sut.Querier(anyMinTime, anyMaxTime)
	require.NoError(t, err, "should return no error")

	ctx := context.Background()
	sorted := true
	hints := &storage.SelectHints{}
	matchers := []*labels.Matcher{
		{Type: labels.MatchEqual,
			Name:  "somename",
			Value: "somevalforlabel"},
		{Type: labels.MatchEqual,
			Name:  "somename2",
			Value: "somevalforlabel2"},
	}

	mergedSeries := querier.Select(ctx, sorted, hints, matchers...)
	graviolaSeriesSet, ok := mergedSeries.(*domain.GraviolaSeriesSet)
	require.True(t, ok, "should be a GraviolaSeriesSet")

	assert.Len(t, graviolaSeriesSet.Series, 2, "should have all the remote storage series")

	for _, serie := range graviolaSeriesSet.Series {
		if reflect.DeepEqual(serie.Labels().Map(), labels.FromStrings("label1", "val1").Map()) {
			assert.Equal(t, []model.SamplePair{{Timestamp: 5819, Value: 5.9}}, serie.Datapoints,
				"should have returned the correct series datapoints")
		} else {
			assert.Equal(t, []model.SamplePair{{Timestamp: 5999, Value: 5.1}}, serie.Datapoints,
				"should have returned the correct series datapoints")
		}
	}
}

func TestConcurrentSelects(t *testing.T) {
	goroutinesTotal := 10

	time1 := rand.Int()
	value1 := rand.Float64()
	samplePair1 := model.SamplePair{Timestamp: model.Time(time1), Value: model.SampleValue(value1)}

	mockStorage1 := &mocks.RemoteStorageMock{
		SeriesSet: &domain.GraviolaSeriesSet{
			Series: []*domain.GraviolaSeries{
				{Lbs: labels.FromStrings("label1", "val1"),
					Datapoints: []model.SamplePair{samplePair1}},
			},
		},
	}

	time2 := rand.Int()
	value2 := rand.Float64()
	samplePair2 := model.SamplePair{Timestamp: model.Time(time2), Value: model.SampleValue(value2)}

	mockStorage2 := &mocks.RemoteStorageMock{
		SeriesSet: &domain.GraviolaSeriesSet{
			Series: []*domain.GraviolaSeries{
				{Lbs: labels.FromStrings("label2", "val2"),
					Datapoints: []model.SamplePair{samplePair2}},
			},
		},
	}

	sut := storageproxy.NewGraviolaStorage(logg, []storage.Querier{mockStorage1, mockStorage2}, defaultMergeStrategy)

	querier, err := sut.Querier(anyMinTime, anyMaxTime)
	require.NoError(t, err, "should return no error")

	ctx := context.Background()
	sorted := true
	hints := &storage.SelectHints{}
	matchers := []*labels.Matcher{
		{Type: labels.MatchEqual,
			Name:  "somename",
			Value: "somevalforlabel"},
		{Type: labels.MatchEqual,
			Name:  "somename2",
			Value: "somevalforlabel2"},
	}

	results := make(chan storage.SeriesSet, goroutinesTotal)
	var wg sync.WaitGroup
	wg.Add(goroutinesTotal)

	go func() {
		wg.Wait()
		close(results)
	}()

	for i := 0; i < goroutinesTotal; i++ {
		go func() {
			results <- querier.Select(ctx, sorted, hints, matchers...)
			wg.Done()
		}()
	}

	counterOfResults := 0
	for res := range results {
		counterOfResults++

		graviolaSeriesSet, ok := res.(*domain.GraviolaSeriesSet)
		require.True(t, ok, "should be a GraviolaSeriesSet")
		assert.Len(t, graviolaSeriesSet.Series, 2, "should have all the remote storage series")

		for _, serie := range graviolaSeriesSet.Series {
			if reflect.DeepEqual(serie.Labels().Map(), labels.FromStrings("label1", "val1").Map()) {
				assert.Equal(t, []model.SamplePair{samplePair1}, serie.Datapoints,
					"should have returned the correct series datapoints")
			} else {
				assert.Equal(t, []model.SamplePair{samplePair2}, serie.Datapoints,
					"should have returned the correct series datapoints")
			}
		}
	}

	assert.Equal(t, goroutinesTotal, counterOfResults, "should have returned all results")
}

func TestConcurrentSelectsWithDifferentAnswers(t *testing.T) {
	goroutinesTotal := 10
	timestampsGenerated := make([]int, 0)
	valuesGenerated := make([]float64, 0)

	mockStorage1 := &mocks.RemoteStorageMock{
		SelectFn: func(_ context.Context, _ bool, _ *storage.SelectHints, _ ...*labels.Matcher) storage.SeriesSet {

			time := rand.Int() //TODO: extract this logic to a function
			for {
				if !slices.Contains(timestampsGenerated, time) {
					break
				}
				time = rand.Int()
			}

			value := rand.Float64()
			for {
				if !slices.Contains(valuesGenerated, value) {
					break
				}
				value = rand.Float64()
			}

			samplePair := model.SamplePair{Timestamp: model.Time(time), Value: model.SampleValue(value)}

			return &domain.GraviolaSeriesSet{
				Series: []*domain.GraviolaSeries{
					{Lbs: labels.FromStrings("label1", "val1"),
						Datapoints: []model.SamplePair{samplePair}},
				},
			}
		},
	}

	mockStorage2 := &mocks.RemoteStorageMock{
		SelectFn: func(_ context.Context, _ bool, _ *storage.SelectHints, _ ...*labels.Matcher) storage.SeriesSet {

			time := rand.Int()
			for {
				if !slices.Contains(timestampsGenerated, time) {
					break
				}
				time = rand.Int()

			}

			value := rand.Float64()
			for {
				if !slices.Contains(valuesGenerated, value) {
					break
				}
				value = rand.Float64()
			}

			samplePair := model.SamplePair{Timestamp: model.Time(time), Value: model.SampleValue(value)}

			return &domain.GraviolaSeriesSet{
				Series: []*domain.GraviolaSeries{
					{Lbs: labels.FromStrings("label2", "val2"),
						Datapoints: []model.SamplePair{samplePair}},
				},
			}
		},
	}

	sut := storageproxy.NewGraviolaStorage(logg, []storage.Querier{mockStorage1, mockStorage2}, defaultMergeStrategy)

	querier, err := sut.Querier(0, 6000)
	require.NoError(t, err, "should not return error")

	ctx := context.Background()
	sorted := true
	hints := &storage.SelectHints{}
	matchers := []*labels.Matcher{
		{Type: labels.MatchEqual,
			Name:  "somename",
			Value: "somevalforlabel"},
		{Type: labels.MatchEqual,
			Name:  "somename2",
			Value: "somevalforlabel2"},
	}

	results := make(chan storage.SeriesSet, goroutinesTotal)
	var wg sync.WaitGroup
	wg.Add(goroutinesTotal)

	go func() {
		wg.Wait()
		close(results)
	}()

	for i := 0; i < goroutinesTotal; i++ {
		go func() {
			results <- querier.Select(ctx, sorted, hints, matchers...)
			wg.Done()
		}()
	}

	counterOfResults := 0
	for res := range results {
		counterOfResults++

		graviolaSeriesSet, ok := res.(*domain.GraviolaSeriesSet)
		assert.True(t, ok, "should be a GraviolaSeriesSet")
		assert.Len(t, graviolaSeriesSet.Series, 2, "should have all the remote storage series")

		for _, serie := range graviolaSeriesSet.Series {
			assert.Len(t, serie.Datapoints, 1, "should have all datapoints")
		}
	}

	assert.Equal(t, goroutinesTotal, counterOfResults, "should have returned all results")
}
