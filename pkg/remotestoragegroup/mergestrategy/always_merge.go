package mergestrategy

import (
	"slices"

	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
)

// A merge strategy where all series are always merged together. If it finds the same timestamp,
// then it will keep the first entry found
type AlwaysMergeStrategy struct{}

func NewAlwaysMergeStrategy() *AlwaysMergeStrategy {
	return &AlwaysMergeStrategy{}
}

// The series inside each seriesSet need to be ordered for this to work
func (merger *AlwaysMergeStrategy) Merge(seriesSets []storage.SeriesSet) storage.SeriesSet {
	if len(seriesSets) == 0 {
		return storage.NoopSeriesSet()
	}

	if len(seriesSets) == 1 {
		return seriesSets[0]
	}

	graviolaSeries := keepOnlyGraviolaSeries(seriesSets)

	if len(graviolaSeries) != 0 {
		slices.SortFunc(graviolaSeries, func(a, b *domain.GraviolaSeries) int {
			return labels.Compare(a.Labels(), b.Labels())
		})
	}
	mergedSeries := make([]*domain.GraviolaSeries, 0, len(graviolaSeries))

	var currentSeries *domain.GraviolaSeries = nil
	for _, serie := range graviolaSeries {
		serie := serie

		if currentSeries == nil {
			currentSeries = serie
			continue
		}

		if labels.Equal(currentSeries.Lbs, serie.Lbs) {
			currentSeries.Datapoints = append(currentSeries.Datapoints, serie.Datapoints...)
		} else {
			mergedSeries = append(mergedSeries, currentSeries)
			currentSeries = serie
		}
	}

	if currentSeries != nil {
		mergedSeries = append(mergedSeries, currentSeries)
	}

	for _, serie := range mergedSeries {

		if serie != nil && len(serie.Datapoints) > 0 {
			slices.SortFunc(serie.Datapoints, func(a, b model.SamplePair) int {
				if a.Timestamp.After(b.Timestamp) {
					return 1
				} else if b.Timestamp.After(a.Timestamp) {
					return -1
				}
				return 0
			})

			removeDuplicatedTimestamps(serie)
		}
	}

	annots := mergeAnnotations(seriesSets)
	erro := joinErrors(seriesSets)

	return &domain.GraviolaSeriesSet{
		Series: mergedSeries,
		Annots: *annots,
		Erro:   erro,
	}
}

func removeDuplicatedTimestamps(serie *domain.GraviolaSeries) {
	curTimestamp := model.Time(0)
	dedupedDatapoints := make([]model.SamplePair, 0, len(serie.Datapoints))

	for _, datapoint := range serie.Datapoints {
		if curTimestamp == 0 {
			curTimestamp = datapoint.Timestamp
			dedupedDatapoints = append(dedupedDatapoints, datapoint)
			continue
		}

		if curTimestamp != datapoint.Timestamp {
			dedupedDatapoints = append(dedupedDatapoints, datapoint)
			curTimestamp = datapoint.Timestamp
		}
	}

	serie.Datapoints = dedupedDatapoints
}
