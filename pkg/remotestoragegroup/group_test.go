package remotestoragegroup_test

import (
	"context"
	"testing"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/jademcosta/graviola/pkg/graviolalog"
	"github.com/jademcosta/graviola/pkg/remotestoragegroup"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/stretchr/testify/assert"
)

var logg *graviolalog.Logger = graviolalog.NewLogger(config.LogConfig{Level: "error"})

func TestClose(t *testing.T) {

	sut := remotestoragegroup.NewGroup(logg, "any name", make([]storage.Querier, 0))

	assert.Nil(t, sut.Close(), "should always return nil")
}

func TestMergesMultipleRemoteStorageResults(t *testing.T) {

	seriesSet1 := &domain.GraviolaSeriesSet{
		Series: []*domain.GraviolaSeries{
			{Lbs: labels.FromStrings("yyyyy", "zzzzz"),
				Ts:   []int64{123456, 123486},
				Vals: []float64{9.9, 11.11}},
			{Lbs: labels.FromStrings("label1", "value1", "labelx", "valuey"),
				Ts:   []int64{123, 456},
				Vals: []float64{1.23, 4.56},
			},
		},
	}

	seriesSet2 := &domain.GraviolaSeriesSet{
		Series: []*domain.GraviolaSeries{
			{Lbs: labels.FromStrings("yyyyy", "zzzzz"),
				Ts:   []int64{789, 101112},
				Vals: []float64{7.89, 10.1112}},
			{Lbs: labels.FromStrings("aaaaa", "zzzzz"),
				Ts:   []int64{131415, 161718},
				Vals: []float64{13.1415, 16.1718}},
		},
	}

	remotes := make([]storage.Querier, 0)
	remotes = append(remotes, &RemoteStorageMock{
		SeriesSet: seriesSet1,
	})
	remotes = append(remotes, &RemoteStorageMock{
		SeriesSet: seriesSet2,
	})

	sut := remotestoragegroup.NewGroup(logg, "any name", remotes)

	result := sut.Select(context.Background(), true, &storage.SelectHints{})

	assert.Panics(t, func() { result.At() }, "should panic if At() is called without Next() first")

	assert.True(t, result.Next(), "should have Next")
	series := result.At()
	assert.Equal(t, "{aaaaa=\"zzzzz\"}", series.Labels().String(), "should return the correct series")

	assert.True(t, result.Next(), "should have Next")
	series = result.At()
	assert.Equal(t, "{label1=\"value1\", labelx=\"valuey\"}", series.Labels().String(), "should return the correct series")

	assert.True(t, result.Next(), "should have Next")
	series = result.At()
	assert.Equal(t, "{yyyyy=\"zzzzz\"}", series.Labels().String(), "should return the correct series")

	assert.False(t, result.Next(), "should NOT have Next")
	assert.Panics(t, func() { result.At() }, "should panic if At() is called after Next() returns false")
}
