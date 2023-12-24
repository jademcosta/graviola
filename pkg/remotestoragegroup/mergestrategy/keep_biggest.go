package mergestrategy

import (
	"slices"

	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
)

// A merge strategy where the series with more datapoints is completely kept,
// while the others are completely discarded. In case of series with the same size,
// the first one on the ordering is kept.
type KeepBiggestMergeStrategy struct{}

func NewKeepBiggestMergeStrategy() *KeepBiggestMergeStrategy {
	return &KeepBiggestMergeStrategy{}
}

// The series inside each seriesSet need to be ordered for this to work
func (merger *KeepBiggestMergeStrategy) Merge(seriesSets []storage.SeriesSet) storage.SeriesSet {
	if len(seriesSets) == 0 {
		return storage.NoopSeriesSet()
	}

	if len(seriesSets) == 1 {
		return seriesSets[0]
	}

	graviolaSeries := keepOnlyGraviolaSeries(seriesSets)

	if len(graviolaSeries) == 0 {
		return storage.NoopSeriesSet()
	}

	slices.SortFunc(graviolaSeries, func(a, b *domain.GraviolaSeries) int {
		return labels.Compare(a.Labels(), b.Labels())
	})

	mergedSeries := make([]*domain.GraviolaSeries, 0, len(graviolaSeries))

	var currentSeries *domain.GraviolaSeries = nil
	for _, serie := range graviolaSeries {
		serie := serie

		if currentSeries == nil {
			currentSeries = serie
			continue
		}

		if labels.Equal(currentSeries.Lbs, serie.Lbs) {
			if len(serie.Datapoints) > len(currentSeries.Datapoints) {
				currentSeries = serie
			}
		} else {
			mergedSeries = append(mergedSeries, currentSeries)
			currentSeries = serie
		}
	}

	mergedSeries = append(mergedSeries, currentSeries)

	return &domain.GraviolaSeriesSet{
		Series: mergedSeries,
	}
}
