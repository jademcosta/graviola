package queryfailurestrategy_test

import (
	"errors"
	"testing"
	"time"

	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/jademcosta/graviola/pkg/remotestoragegroup/queryfailurestrategy"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var sut = &queryfailurestrategy.PartialResponseStrategy{}

func TestForLabels(t *testing.T) {
	t.Run("does nothing when error is nil", func(t *testing.T) {
		t.Parallel()

		lbls := []string{"a", "a", "b", "c"}
		lblsResponse, err := sut.ForLabels(lbls, nil)
		require.NoError(t, err, "should return no error")
		assert.Equal(t, lbls, lblsResponse, "should return all labels")
		assert.Equal(t, &lbls, &lblsResponse, "should return the same labels")
	})

	t.Run("does nothing when no labels are passed to it", func(t *testing.T) {
		t.Parallel()

		lbls := []string{}
		lblsResponse, err := sut.ForLabels(lbls, nil)
		require.NoError(t, err, "should return no error")
		assert.Equal(t, lbls, lblsResponse, "should return all labels")
		assert.Equal(t, &lbls, &lblsResponse, "should return the same labels")
		assert.Empty(t, lblsResponse, "should return the same labels")

		error1 := errors.New("some err")

		lblsResponse, err = sut.ForLabels(lbls, error1)
		assert.Equal(t, error1, err, "should return the same error")
		assert.Equal(t, lbls, lblsResponse, "should return all labels")
		assert.Equal(t, &lbls, &lblsResponse, "should return the same labels")
		assert.Empty(t, lblsResponse, "should return the same labels")

		lblsResponse, err = sut.ForLabels(nil, nil)
		require.NoError(t, err, "should return no error")
		assert.Nil(t, lblsResponse, "should return all labels")
		assert.Empty(t, lblsResponse, "should return the same labels")
	})

	t.Run("removes the error when there is at least one valid answer", func(t *testing.T) {
		t.Parallel()

		lbls := []string{"a"}
		error1 := errors.New("some error")

		lblsResponse, err := sut.ForLabels(lbls, error1)
		require.NoError(t, err, "should return no error")
		assert.Equal(t, lbls, lblsResponse, "should return all labels")
		assert.Equal(t, &lbls, &lblsResponse, "should return the same labels")
	})
}

func TestForSeriesSet(t *testing.T) {
	t.Run("does nothing when error is nil", func(t *testing.T) {
		t.Parallel()

		sSet := &domain.GraviolaSeriesSet{
			Erro: nil,
			Series: []*domain.GraviolaSeries{{
				Datapoints: []model.SamplePair{
					{Timestamp: model.Time(time.Now().Unix()), Value: 10},
				}},
			},
		}

		response := sut.ForSeriesSet(sSet)
		require.NoError(t, response.Err(), "should return no error")
		assert.Equal(t, sSet, response, "should return all seriesSet data")
	})

	t.Run("does nothing when series set has no series", func(t *testing.T) {
		t.Parallel()

		sSet := &domain.GraviolaSeriesSet{
			Erro: nil,
		}

		response := sut.ForSeriesSet(sSet)
		require.NoError(t, response.Err(), "should return no error")
		assert.Equal(t, sSet, response, "should return the same series set")
		assert.False(t, response.Next(), "should return the same series set")

		sSet = &domain.GraviolaSeriesSet{
			Erro:   nil,
			Series: []*domain.GraviolaSeries{},
		}

		response = sut.ForSeriesSet(sSet)
		require.NoError(t, response.Err(), "should return no error")
		assert.Equal(t, sSet, response, "should return the same series set")
		assert.False(t, response.Next(), "should return the same series set")

		error1 := errors.New("some error")

		sSet = &domain.GraviolaSeriesSet{
			Erro:   error1,
			Series: []*domain.GraviolaSeries{},
		}

		response = sut.ForSeriesSet(sSet)
		require.Error(t, response.Err(), "should return the same error")
		assert.Equal(t, sSet, response, "should return the same series set")
		assert.False(t, response.Next(), "should return the same series set")
	})

	t.Run("removes the error if it has at least 1 series", func(t *testing.T) {
		t.Parallel()

		error1 := errors.New("some error")

		sSet := &domain.GraviolaSeriesSet{
			Erro: error1,
			Series: []*domain.GraviolaSeries{{
				Datapoints: []model.SamplePair{
					{Timestamp: model.Time(time.Now().Unix()), Value: 10},
				}},
			},
		}

		response := sut.ForSeriesSet(sSet)
		require.NoError(t, response.Err(), "should return no error")
		assert.Equal(t, sSet, response, "should return the same series set (it is a pointer)")
		assert.True(t, response.Next(), "should return the same series set")
	})
}
