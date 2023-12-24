package mergestrategy_test

import (
	"testing"

	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/jademcosta/graviola/pkg/remotestoragegroup/mergestrategy"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/assert"
)

func TestTheAlwaysMerrgeMergeMethod(t *testing.T) {

	t.Run("with a single set", func(t *testing.T) {
		series1 := []*domain.GraviolaSeries{
			{Lbs: labels.FromStrings("x", "value1"),
				Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 1.1}, {Timestamp: 1703379286017, Value: 1.2}}},
		}

		seriesSet := []*domain.GraviolaSeriesSet{
			{Series: series1},
		}

		sut := mergestrategy.NewAlwaysMergeStrategy()

		resp := sut.Merge(cast(seriesSet))

		parsedSet, ok := resp.(*domain.GraviolaSeriesSet)
		assert.True(t, ok, "parsing should work")

		expected := &domain.GraviolaSeriesSet{
			Series: []*domain.GraviolaSeries{
				{Lbs: labels.FromStrings("x", "value1"),
					Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 1.1}, {Timestamp: 1703379286017, Value: 1.2}}},
			},
		}

		assert.Equal(t, expected, parsedSet, "should match")
		equal(t, expected, parsedSet)
	})

	t.Run("with simple data", func(t *testing.T) {
		series1 := []*domain.GraviolaSeries{
			{Lbs: labels.FromStrings("x", "value1"),
				Datapoints: []model.SamplePair{{Timestamp: 1703379256018, Value: 11.0}}},
			{Lbs: labels.FromStrings("label1", "value1"),
				Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 2.1}}},
		}

		series2 := []*domain.GraviolaSeries{
			{Lbs: labels.FromStrings("x", "value1"),
				Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 1.1}, {Timestamp: 1703379286017, Value: 1.2}}},
			{Lbs: labels.FromStrings("labelaaaa", "valueaaaa"),
				Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 0.1}}},
		}

		seriesSet := []*domain.GraviolaSeriesSet{
			{Series: series2},
			{Series: series1},
		}

		sut := mergestrategy.NewAlwaysMergeStrategy()

		resp := sut.Merge(cast(seriesSet))

		parsedSet, ok := resp.(*domain.GraviolaSeriesSet)
		assert.True(t, ok, "parsing should work")

		expected := &domain.GraviolaSeriesSet{
			Series: []*domain.GraviolaSeries{
				{Lbs: labels.FromStrings("label1", "value1"),
					Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 2.1}}},
				{Lbs: labels.FromStrings("labelaaaa", "valueaaaa"),
					Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 0.1}}},
				{Lbs: labels.FromStrings("x", "value1"),
					Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 1.1}, {Timestamp: 1703379256018, Value: 11.0}, {Timestamp: 1703379286017, Value: 1.2}}},
			},
		}

		assert.Equal(t, expected, parsedSet, "should match")
		equal(t, expected, parsedSet)
	})

	t.Run("when conflicting data has the same timesstamp, it picks the first one", func(t *testing.T) {
		series1 := []*domain.GraviolaSeries{
			{Lbs: labels.FromStrings("x", "value1"),
				Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 1.1}, {Timestamp: 1703379286017, Value: 1.2}}},
		}

		series2 := []*domain.GraviolaSeries{
			{Lbs: labels.FromStrings("x", "value1"),
				Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 11.0}, {Timestamp: 1703379286017, Value: 11.2}}},
		}

		seriesSet := []*domain.GraviolaSeriesSet{
			{Series: series1},
			{Series: series2},
		}

		sut := mergestrategy.NewAlwaysMergeStrategy()

		resp := sut.Merge(cast(seriesSet))

		parsedSet, ok := resp.(*domain.GraviolaSeriesSet)
		assert.True(t, ok, "parsing should work")

		expected := &domain.GraviolaSeriesSet{
			Series: []*domain.GraviolaSeries{
				{Lbs: labels.FromStrings("x", "value1"),
					Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 1.1}, {Timestamp: 1703379286017, Value: 1.2}}},
			},
		}

		assert.Equal(t, expected, parsedSet, "should match")
		equal(t, expected, parsedSet)
	})

	t.Run("it keeps labels with different values as separate values", func(t *testing.T) {
		series1 := []*domain.GraviolaSeries{
			{Lbs: labels.FromStrings("x", "value1"),
				Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 1.1}, {Timestamp: 1703379286017, Value: 1.2}}},
		}

		series2 := []*domain.GraviolaSeries{
			{Lbs: labels.FromStrings("x", "value1"),
				Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 11.0}, {Timestamp: 1703379286017, Value: 11.2}}},
			{Lbs: labels.FromStrings("x", "value"),
				Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 16.0}, {Timestamp: 1703379286017, Value: 26.2}}},
		}

		seriesSet := []*domain.GraviolaSeriesSet{
			{Series: series1},
			{Series: series2},
		}

		sut := mergestrategy.NewAlwaysMergeStrategy()

		resp := sut.Merge(cast(seriesSet))

		parsedSet, ok := resp.(*domain.GraviolaSeriesSet)
		assert.True(t, ok, "parsing should work")

		expected := &domain.GraviolaSeriesSet{
			Series: []*domain.GraviolaSeries{
				{Lbs: labels.FromStrings("x", "value"),
					Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 16.0}, {Timestamp: 1703379286017, Value: 26.2}}},
				{Lbs: labels.FromStrings("x", "value1"),
					Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 1.1}, {Timestamp: 1703379286017, Value: 1.2}}},
			},
		}

		assert.Equal(t, expected, parsedSet, "should match")
		equal(t, expected, parsedSet)
	})

	t.Run("with multiple conflicts and series, it considers a conflict only if data has the same timestamp", func(t *testing.T) {
		series1 := []*domain.GraviolaSeries{
			{Lbs: labels.FromStrings("x", "value1"),
				Datapoints: []model.SamplePair{{Timestamp: 12, Value: 1.1}, {Timestamp: 15, Value: 1.2}, {Timestamp: 18, Value: 1.3}}},
		}

		series2 := []*domain.GraviolaSeries{
			{Lbs: labels.FromStrings("x", "value1"),
				Datapoints: []model.SamplePair{{Timestamp: 12, Value: 11.1}, {Timestamp: 13, Value: 11.2}}},
			{Lbs: labels.FromStrings("x", "value"),
				Datapoints: []model.SamplePair{{Timestamp: 12, Value: 16.0}, {Timestamp: 13, Value: 26.2}}},
		}

		series3 := []*domain.GraviolaSeries{
			{Lbs: labels.FromStrings("x", "value1"),
				Datapoints: []model.SamplePair{{Timestamp: 12, Value: 21.1}, {Timestamp: 17, Value: 21.2}}},
		}

		series4 := []*domain.GraviolaSeries{
			{Lbs: labels.FromStrings("x", "value1"),
				Datapoints: []model.SamplePair{{Timestamp: 21, Value: 31.1}}},
		}

		seriesSet := []*domain.GraviolaSeriesSet{
			{Series: series1},
			{Series: series2},
			{Series: series3},
			{Series: series4},
		}

		sut := mergestrategy.NewAlwaysMergeStrategy()

		resp := sut.Merge(cast(seriesSet))

		parsedSet, ok := resp.(*domain.GraviolaSeriesSet)
		assert.True(t, ok, "parsing should work")

		expected := &domain.GraviolaSeriesSet{
			Series: []*domain.GraviolaSeries{
				{Lbs: labels.FromStrings("x", "value"),
					Datapoints: []model.SamplePair{{Timestamp: 12, Value: 16.0}, {Timestamp: 13, Value: 26.2}}},
				{Lbs: labels.FromStrings("x", "value1"),
					Datapoints: []model.SamplePair{{Timestamp: 12, Value: 1.1}, {Timestamp: 13, Value: 11.2}, {Timestamp: 15, Value: 1.2}, {Timestamp: 17, Value: 21.2}, {Timestamp: 18, Value: 1.3}, {Timestamp: 21, Value: 31.1}}},
			},
		}

		assert.Equal(t, expected, parsedSet, "should match")
		equal(t, expected, parsedSet)
	})
}
