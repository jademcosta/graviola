package remotestoragegroup_test

import (
	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/graviolalog"
)

var logg *graviolalog.Logger = graviolalog.NewLogger(config.LogConfig{Level: "error"})

// func TestClose(t *testing.T) {

// 	sut := remotestoragegroup.NewGroup(logg, "any name", make([]storage.Querier, 0))

// 	assert.Nil(t, sut.Close(), "should always return nil")
// }

// func TestMergesMultipleRemoteStorageResults(t *testing.T) {

// 	series1 := []*domain.GraviolaSeries{
// 		{Lbs: labels.FromStrings("yyyyy", "zzzzz"),
// 			Datapoints: []model.SamplePair{{Timestamp: 123456, Value: 9.9}, {Timestamp: 123486, Value: 11.11}}},
// 		{Lbs: labels.FromStrings("label1", "value1", "labelx", "valuey"),
// 			Datapoints: []model.SamplePair{{Timestamp: 123, Value: 1.23}, {Timestamp: 456, Value: 4.56}}},
// 	}

// 	seriesSet1 := &domain.GraviolaSeriesSet{Series: series1}

// 	series2 := []*domain.GraviolaSeries{
// 		{Lbs: labels.FromStrings("yyyyy", "zzzzz"),
// 			Datapoints: []model.SamplePair{{Timestamp: 789, Value: 7.89}, {Timestamp: 101112, Value: 10.1112}}},
// 		{Lbs: labels.FromStrings("aaaaa", "zzzzz"),
// 			Datapoints: []model.SamplePair{{Timestamp: 131415, Value: 13.1415}, {Timestamp: 161718, Value: 16.1718}}},
// 	}

// 	seriesSet2 := &domain.GraviolaSeriesSet{Series: series2}

// 	remotes := make([]storage.Querier, 0)
// 	remotes = append(remotes, &RemoteStorageMock{
// 		SeriesSet: seriesSet1,
// 	})
// 	remotes = append(remotes, &RemoteStorageMock{
// 		SeriesSet: seriesSet2,
// 	})

// 	sut := remotestoragegroup.NewGroup(logg, "any name", remotes)

// 	result := sut.Select(context.Background(), true, &storage.SelectHints{})

// 	assert.Panics(t, func() { result.At() }, "should panic if At() is called without Next() first")

// 	assert.True(t, result.Next(), "should have Next")
// 	series := result.At()
// 	assert.Equal(t, "{aaaaa=\"zzzzz\"}", series.Labels().String(), "should return the correct series")

// 	assert.True(t, result.Next(), "should have Next")
// 	series = result.At()
// 	assert.Equal(t, "{label1=\"value1\", labelx=\"valuey\"}", series.Labels().String(), "should return the correct series")

// 	assert.True(t, result.Next(), "should have Next")
// 	series = result.At()
// 	assert.Equal(t, "{yyyyy=\"zzzzz\"}", series.Labels().String(), "should return the correct series")

// 	assert.False(t, result.Next(), "should NOT have Next")
// 	assert.Panics(t, func() { result.At() }, "should panic if At() is called after Next() returns false")
// }
