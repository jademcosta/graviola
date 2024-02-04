package mergestrategy_test

import (
	"errors"
	"testing"

	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/jademcosta/graviola/pkg/remotestoragegroup/mergestrategy"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/util/annotations"
	"github.com/stretchr/testify/assert"
)

func TestTheMergeMethod(t *testing.T) {

	t.Run("with a single set", func(t *testing.T) {
		series1 := []*domain.GraviolaSeries{
			{Lbs: labels.FromStrings("x", "value1"),
				Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 1.1}, {Timestamp: 1703379286017, Value: 1.2}}},
		}

		seriesSet := []*domain.GraviolaSeriesSet{
			{Series: series1},
		}

		sut := mergestrategy.NewKeepBiggestMergeStrategy()

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
		assert.NoError(t, resp.Err(), "should return no error")
	})

	t.Run("with simple data", func(t *testing.T) {
		series1 := []*domain.GraviolaSeries{
			{Lbs: labels.FromStrings("x", "value1"),
				Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 11.0}}},
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

		sut := mergestrategy.NewKeepBiggestMergeStrategy()

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
					Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 1.1}, {Timestamp: 1703379286017, Value: 1.2}}},
			},
			Annots: *annotations.New(),
		}

		assert.Equal(t, expected, parsedSet, "should match")
		equal(t, expected, parsedSet)
		assert.NoError(t, resp.Err(), "should return no error")
	})

	t.Run("when conflicting data has the same size, it picks the first one", func(t *testing.T) {
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

		sut := mergestrategy.NewKeepBiggestMergeStrategy()

		resp := sut.Merge(cast(seriesSet))

		parsedSet, ok := resp.(*domain.GraviolaSeriesSet)
		assert.True(t, ok, "parsing should work")

		expected := &domain.GraviolaSeriesSet{
			Series: []*domain.GraviolaSeries{
				{Lbs: labels.FromStrings("x", "value1"),
					Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 1.1}, {Timestamp: 1703379286017, Value: 1.2}}},
			},
			Annots: *annotations.New(),
		}

		assert.Equal(t, expected, parsedSet, "should match")
		equal(t, expected, parsedSet)
		assert.NoError(t, resp.Err(), "should return no error")
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

		sut := mergestrategy.NewKeepBiggestMergeStrategy()

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
			Annots: *annotations.New(),
		}

		assert.Equal(t, expected, parsedSet, "should match")
		equal(t, expected, parsedSet)
		assert.NoError(t, resp.Err(), "should return no error")
	})

	t.Run("multiple conflicts, the biggest keeps being kept", func(t *testing.T) {
		series1 := []*domain.GraviolaSeries{
			{Lbs: labels.FromStrings("x", "value1"),
				Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 1.1}, {Timestamp: 1703379286017, Value: 1.2}, {Timestamp: 1703379316017, Value: 1.3}}},
		}

		series2 := []*domain.GraviolaSeries{
			{Lbs: labels.FromStrings("x", "value1"),
				Datapoints: []model.SamplePair{{Timestamp: 703379256017, Value: 11.1}, {Timestamp: 703379286017, Value: 11.2}}},
			{Lbs: labels.FromStrings("x", "value"),
				Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 16.0}, {Timestamp: 1703379286017, Value: 26.2}}},
		}

		series3 := []*domain.GraviolaSeries{
			{Lbs: labels.FromStrings("x", "value1"),
				Datapoints: []model.SamplePair{{Timestamp: 703379256017, Value: 21.1}, {Timestamp: 703379286017, Value: 21.2}}},
		}

		series4 := []*domain.GraviolaSeries{
			{Lbs: labels.FromStrings("x", "value1"),
				Datapoints: []model.SamplePair{{Timestamp: 1703389256017, Value: 31.1}}},
		}

		seriesSet := []*domain.GraviolaSeriesSet{
			{Series: series1},
			{Series: series2},
			{Series: series3},
			{Series: series4},
		}

		sut := mergestrategy.NewKeepBiggestMergeStrategy()

		resp := sut.Merge(cast(seriesSet))

		parsedSet, ok := resp.(*domain.GraviolaSeriesSet)
		assert.True(t, ok, "parsing should work")

		expected := &domain.GraviolaSeriesSet{
			Series: []*domain.GraviolaSeries{
				{Lbs: labels.FromStrings("x", "value"),
					Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 16.0}, {Timestamp: 1703379286017, Value: 26.2}}},
				{Lbs: labels.FromStrings("x", "value1"),
					Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 1.1}, {Timestamp: 1703379286017, Value: 1.2}, {Timestamp: 1703379316017, Value: 1.3}}},
			},
			Annots: *annotations.New(),
		}

		assert.Equal(t, expected, parsedSet, "should match")
		equal(t, expected, parsedSet)
		assert.NoError(t, resp.Err(), "should return no error")
	})

	t.Run("merges Annotations of all SeriesSets", func(t *testing.T) {

		err1 := errors.New("some random error")
		err2 := errors.New("expected error")

		seriesSet := []*domain.GraviolaSeriesSet{
			{Annots: annotations.New().Add(err1)},
			{Annots: annotations.New().Add(err2)},
			{Annots: annotations.New().Add(err2)},
			{Annots: annotations.New().Add(err2)},
		}

		sut := mergestrategy.NewKeepBiggestMergeStrategy()

		resp := sut.Merge(cast(seriesSet))

		parsedSet, ok := resp.(*domain.GraviolaSeriesSet)
		assert.True(t, ok, "parsing should work")

		assert.Len(t, parsedSet.Annots, 2, "should keep all annotations")
		assert.Contains(t, parsedSet.Warnings(), err1.Error(), "should contain the annotation")
		assert.Contains(t, parsedSet.Warnings(), err2.Error(), "should contain the annotation")
		assert.Len(t, parsedSet.Series, 0, "should not create any series out of nothing")
		assert.NoError(t, resp.Err(), "should return no error")
	})

	t.Run("merges errors of all answers", func(t *testing.T) {

		err1 := errors.New("some random error")
		err2 := errors.New("expected error")

		seriesSet := []*domain.GraviolaSeriesSet{
			{Erro: err1},
			{Erro: err2},
			{
				Series: []*domain.GraviolaSeries{
					{Lbs: labels.FromStrings("x", "value"),
						Datapoints: []model.SamplePair{{Timestamp: 1703379256017, Value: 16.0}, {Timestamp: 1703379286017, Value: 26.2}}},
				},
			},
		}

		sut := mergestrategy.NewKeepBiggestMergeStrategy()

		resp := sut.Merge(cast(seriesSet))
		assert.Error(t, resp.Err(), "should return an error")

		parsedSet, ok := resp.(*domain.GraviolaSeriesSet)
		assert.True(t, ok, "parsing should work")

		assert.Len(t, parsedSet.Annots, 0, "should not be responsible for turning errors into annotations")
		assert.Len(t, resp.Warnings(), 0, "should not be responsible for turning errors into annotations")

		assert.Len(t, parsedSet.Series, 1, "should keep the returned series")
		assert.ErrorIs(t, resp.Err(), err1, "should have joined the returned errors")
		assert.ErrorIs(t, resp.Err(), err2, "should have joined the returned errors")
	})
}
