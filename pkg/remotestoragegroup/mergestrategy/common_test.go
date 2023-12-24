package mergestrategy_test

import (
	"testing"

	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/prometheus/prometheus/storage"
	"github.com/stretchr/testify/assert"
)

func cast(seriesSet []*domain.GraviolaSeriesSet) []storage.SeriesSet {
	casted := make([]storage.SeriesSet, 0, len(seriesSet))
	for _, sSet := range seriesSet {
		casted = append(casted, sSet)
	}
	return casted
}

func equal(t *testing.T, setExpected *domain.GraviolaSeriesSet, setResult *domain.GraviolaSeriesSet) {
	assert.Equal(t, len(setExpected.Series), len(setResult.Series), "should have the same length")

	for setExpected.Next() {
		setResult.Next()

		assert.Equal(t, setExpected.At(), setResult.At(), "the series should be equal")
	}
}
