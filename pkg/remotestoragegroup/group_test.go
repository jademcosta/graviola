package remotestoragegroup_test

import (
	"context"
	"log/slog"
	"math/rand"
	"reflect"
	"slices"
	"sync"
	"testing"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/jademcosta/graviola/pkg/graviolalog"
	"github.com/jademcosta/graviola/pkg/remotestoragegroup"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/stretchr/testify/assert"
)

var logg *slog.Logger = graviolalog.NewLogger(config.LogConfig{Level: "error"})

func TestCloseIsSentToRemotes(t *testing.T) {
	mockStorage1 := &RemoteStorageMock{}
	mockStorage2 := &RemoteStorageMock{}

	sut := remotestoragegroup.NewGroup(logg, "any name", []storage.Querier{mockStorage1, mockStorage2})
	sut.Close()

	assert.Equal(t, 1, mockStorage1.closeCalled, "should have called close on wrapper remotes")
	assert.Equal(t, 1, mockStorage2.closeCalled, "should have called close on wrapper remotes")
}

func TestSelect(t *testing.T) {

	mockStorage1 := &RemoteStorageMock{
		SeriesSet: &domain.GraviolaSeriesSet{
			Series: []*domain.GraviolaSeries{
				{Lbs: labels.FromStrings("label1", "val1"),
					Datapoints: []model.SamplePair{{Timestamp: 5819, Value: 5.9}}},
			},
		},
	}
	mockStorage2 := &RemoteStorageMock{
		SeriesSet: &domain.GraviolaSeriesSet{
			Series: []*domain.GraviolaSeries{
				{Lbs: labels.FromStrings("label2", "val2"),
					Datapoints: []model.SamplePair{{Timestamp: 5999, Value: 5.1}}},
			},
		},
	}

	sut := remotestoragegroup.NewGroup(logg, "any name", []storage.Querier{mockStorage1, mockStorage2})

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

	mergedSeries := sut.Select(ctx, sorted, hints, matchers...)
	graviolaSeriesSet, ok := mergedSeries.(*domain.GraviolaSeriesSet)
	assert.True(t, ok, "should be a GraviolaSeriesSet")

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

	mockStorage1 := &RemoteStorageMock{
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

	mockStorage2 := &RemoteStorageMock{
		SeriesSet: &domain.GraviolaSeriesSet{
			Series: []*domain.GraviolaSeries{
				{Lbs: labels.FromStrings("label2", "val2"),
					Datapoints: []model.SamplePair{samplePair2}},
			},
		},
	}

	sut := remotestoragegroup.NewGroup(logg, "any name", []storage.Querier{mockStorage1, mockStorage2})

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
			results <- sut.Select(ctx, sorted, hints, matchers...)
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

	mockStorage1 := &RemoteStorageMock{
		SelectFn: func(ctx context.Context, b bool, sh *storage.SelectHints, m ...*labels.Matcher) storage.SeriesSet {

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

	mockStorage2 := &RemoteStorageMock{
		SelectFn: func(ctx context.Context, b bool, sh *storage.SelectHints, m ...*labels.Matcher) storage.SeriesSet {

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

	sut := remotestoragegroup.NewGroup(logg, "any name", []storage.Querier{mockStorage1, mockStorage2})

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
			results <- sut.Select(ctx, sorted, hints, matchers...)
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
