package domain_test

import (
	"context"
	"io"
	"math/rand"
	"testing"

	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/util/annotations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWarnings(t *testing.T) {

	errors := map[string]error{"contextErr": context.DeadlineExceeded}

	sut := &domain.GraviolaSeriesSet{
		Annots: errors,
	}

	assert.Equal(t, annotations.Annotations(errors), sut.Warnings(), "should return the annotations")

	annots := annotations.Annotations(map[string]error{"contextErr": context.DeadlineExceeded, "eof": io.EOF})

	sut = &domain.GraviolaSeriesSet{
		Annots: annots,
	}

	assert.Equal(t, annots, sut.Warnings(), "should return the annotations")
}

func TestNextReturnsFalseWhenSeriesIsEmpty(t *testing.T) {
	sut := &domain.GraviolaSeriesSet{
		Annots: map[string]error{"contextErr": context.DeadlineExceeded},
	}

	assert.False(t, sut.Next(), "should return false")
}

func TestNextReturnsTrueWhileThereIsDataAvailable(t *testing.T) {
	sut := &domain.GraviolaSeriesSet{
		Series: []*domain.GraviolaSeries{{Datapoints: []model.SamplePair{{Timestamp: 1, Value: 0.0}, {Timestamp: 2, Value: 0.0}, {Timestamp: 3, Value: 0.0}}}},
	}

	assert.True(t, sut.Next(), "should return true")
	assert.False(t, sut.Next(), "should return false")

	sut = &domain.GraviolaSeriesSet{
		Series: []*domain.GraviolaSeries{
			{Datapoints: []model.SamplePair{
				{Timestamp: 1, Value: 0.0}, {Timestamp: 2, Value: 0.0}, {Timestamp: 3, Value: 0.0}}},
			{Datapoints: []model.SamplePair{
				{Timestamp: 4, Value: 0.0}, {Timestamp: 5, Value: 0.0}, {Timestamp: 6, Value: 0.0}}},
			{Datapoints: []model.SamplePair{
				{Timestamp: 7, Value: 0.0}, {Timestamp: 8, Value: 0.0}, {Timestamp: 9, Value: 0.0}}},
		},
	}

	assert.True(t, sut.Next(), "should return true")
	assert.True(t, sut.Next(), "should return true")
	assert.True(t, sut.Next(), "should return true")
	assert.False(t, sut.Next(), "should return false")
}

func TestAtAndNextAreConnectedAndKeepTheProvidedOrder(t *testing.T) {

	amountOfSeries := rand.Intn(1000)
	gravSeries := make([]*domain.GraviolaSeries, 0, amountOfSeries)

	for i := 0; i < amountOfSeries; i++ {
		gravSeries = append(gravSeries, &domain.GraviolaSeries{Datapoints: []model.SamplePair{{Timestamp: model.Time(rand.Intn(1000)), Value: 0.0}}})
	}

	sut := &domain.GraviolaSeriesSet{
		Series: gravSeries,
	}

	assert.Len(t, sut.Series, amountOfSeries, "should contain the correct amount of series")

	for _, serie := range gravSeries {
		assert.True(t, sut.Next(), "should return true")
		atSerie := sut.At()

		assert.Equal(t, serie, atSerie, "should return the correct serie (ordered)")
	}

	assert.False(t, sut.Next(), "should return false")
}

func TestReturnsErrorIfItNotNil(t *testing.T) {
	sut := &domain.GraviolaSeriesSet{}
	require.NoError(t, sut.Err(), "should return nil when no error is set")

	sut = &domain.GraviolaSeriesSet{Erro: context.DeadlineExceeded}
	assert.Equal(t, context.DeadlineExceeded, sut.Err(), "should return the error when it is set")
}
