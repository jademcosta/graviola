package remotestoragegroup_test

import (
	"context"
	"slices"
	"testing"

	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/jademcosta/graviola/pkg/remotestoragegroup"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/stretchr/testify/assert"
)

type mockMergeStrategy struct {
	toReturn   storage.SeriesSet
	calledWith [][]storage.SeriesSet
}

func (mock *mockMergeStrategy) Merge(seriesSets []storage.SeriesSet) storage.SeriesSet {
	mock.calledWith = append(mock.calledWith, seriesSets)

	if mock.toReturn != nil {
		return mock.toReturn
	}

	return storage.NoopSeriesSet()
}

func TestPanicsIfMergeStrategyIsNil(t *testing.T) {

	assert.Panics(t, func() { remotestoragegroup.NewMergeQuerier([]storage.Querier{}, nil) }, "panics if the merge strategy provided is nil")
}

func TestCloseIsAlwaysNil(t *testing.T) {

	sut := remotestoragegroup.NewMergeQuerier(nil, &mockMergeStrategy{})

	assert.Nil(t, sut.Close(), "Close() should always return nil")
}

func TestClosePropagatesToQueriers(t *testing.T) {

	querier1 := &RemoteStorageMock{}
	querier2 := &RemoteStorageMock{}

	sut := remotestoragegroup.NewMergeQuerier([]storage.Querier{querier1, querier2}, &mockMergeStrategy{})

	assert.Nil(t, sut.Close(), "Close() should always return nil")

	assert.Equal(t, 1, querier1.closeCalled, "should have calledd close on wrapped querier")
	assert.Equal(t, 1, querier2.closeCalled, "should have calledd close on wrapped querier")
}

func TestSelectPropagatesToQueriers(t *testing.T) {

	querier1 := &RemoteStorageMock{SeriesSet: &domain.GraviolaSeriesSet{}}
	querier2 := &RemoteStorageMock{SeriesSet: &domain.GraviolaSeriesSet{}}
	querier3 := &RemoteStorageMock{SeriesSet: &domain.GraviolaSeriesSet{}}

	sut := remotestoragegroup.NewMergeQuerier([]storage.Querier{querier1, querier2, querier3}, &mockMergeStrategy{})
	matchers := []*labels.Matcher{{Type: labels.MatchEqual, Name: "somename", Value: "someValue"}}
	hints := &storage.SelectHints{Start: 123456}

	sut.Select(context.Background(), true, hints, matchers...)

	assert.Equal(t, []bool{true}, querier1.calledWithSortSeries, "should have called with the same sortSeries value")
	assert.Equal(t, []bool{true}, querier2.calledWithSortSeries, "should have called with the same sortSeries value")
	assert.Equal(t, []bool{true}, querier3.calledWithSortSeries, "should have called with the same sortSeries value")

	assert.Equal(t, [][]*labels.Matcher{matchers}, querier1.calledWithMatchers, "should have called with the same matchers value")
	assert.Equal(t, [][]*labels.Matcher{matchers}, querier2.calledWithMatchers, "should have called with the same matchers value")
	assert.Equal(t, [][]*labels.Matcher{matchers}, querier3.calledWithMatchers, "should have called with the same matchers value")

	assert.Equal(t, []*storage.SelectHints{hints}, querier1.calledWithHints, "should have called with the same hints value")
	assert.Equal(t, []*storage.SelectHints{hints}, querier2.calledWithHints, "should have called with the same hints value")
	assert.Equal(t, []*storage.SelectHints{hints}, querier3.calledWithHints, "should have called with the same hints value")
}

func TestCallsTheMergeStrategyWithReturnOfWrappedQueriers(t *testing.T) {

	seriesSet1 := &domain.GraviolaSeriesSet{Series: []*domain.GraviolaSeries{
		{Lbs: labels.FromStrings("key1", "val1"),
			Datapoints: []model.SamplePair{{Timestamp: 12345, Value: 1.1}}},
	}}
	querier1 := &RemoteStorageMock{SeriesSet: seriesSet1}

	seriesSet2 := &domain.GraviolaSeriesSet{Series: []*domain.GraviolaSeries{
		{Lbs: labels.FromStrings("key1", "val1"),
			Datapoints: []model.SamplePair{{Timestamp: 123456, Value: 12.1}}},
	}}
	querier2 := &RemoteStorageMock{SeriesSet: seriesSet2}

	seriesSet3 := &domain.GraviolaSeriesSet{Series: []*domain.GraviolaSeries{
		{Lbs: labels.FromStrings("key1", "val1"),
			Datapoints: []model.SamplePair{{Timestamp: 1234567, Value: 32.3}}},
	}}
	querier3 := &RemoteStorageMock{SeriesSet: seriesSet3}

	strategy := &mockMergeStrategy{}
	sut := remotestoragegroup.NewMergeQuerier([]storage.Querier{querier1, querier2, querier3}, strategy)

	sut.Select(context.Background(), true, &storage.SelectHints{})

	assert.Len(t, strategy.calledWith[0], 3, "should have called with the result of the Queries select")

	assert.True(t, slices.Contains(strategy.calledWith[0], storage.SeriesSet(seriesSet1)), "should contain all SeriesSet returned")
	assert.True(t, slices.Contains(strategy.calledWith[0], storage.SeriesSet(seriesSet2)), "should contain all SeriesSet returned")
	assert.True(t, slices.Contains(strategy.calledWith[0], storage.SeriesSet(seriesSet3)), "should contain all SeriesSet returned")
}

func TestReturnsWhateverTheMergeStrategyReturns(t *testing.T) {
	emptySeriesSet := &domain.GraviolaSeriesSet{}
	querier1 := &RemoteStorageMock{
		SeriesSet: emptySeriesSet,
	}

	querier2 := &RemoteStorageMock{
		SeriesSet: emptySeriesSet,
	}

	expected := &domain.GraviolaSeriesSet{
		Series: []*domain.GraviolaSeries{{Lbs: labels.FromStrings("lbl1", "val1"), Datapoints: []model.SamplePair{{Timestamp: 123, Value: 1.1}}}},
	}

	mergeStrategy := &mockMergeStrategy{toReturn: expected}
	sut := remotestoragegroup.NewMergeQuerier([]storage.Querier{querier1, querier2}, mergeStrategy)

	result := sut.Select(context.Background(), true, &storage.SelectHints{})
	assert.Equal(t, expected, result, "should return what the Merge() call returns")
}

func TestWhenOnlyOneQuerierExistDoesNotCallMerge(t *testing.T) {
	emptySeriesSet := &domain.GraviolaSeriesSet{}
	querier1 := &RemoteStorageMock{
		SeriesSet: emptySeriesSet,
	}

	expected := &domain.GraviolaSeriesSet{
		Series: []*domain.GraviolaSeries{{Lbs: labels.FromStrings("lbl1", "val1"), Datapoints: []model.SamplePair{{Timestamp: 123, Value: 1.1}}}},
	}

	mergeStrategy := &mockMergeStrategy{toReturn: expected}
	sut := remotestoragegroup.NewMergeQuerier([]storage.Querier{querier1}, mergeStrategy)

	sut.Select(context.Background(), true, &storage.SelectHints{})

	assert.Len(t, mergeStrategy.calledWith, 0, "should not call Merge() when only 1 querier exists")
}
