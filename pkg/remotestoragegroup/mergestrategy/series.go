package mergestrategy

import (
	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/util/annotations"
)

func keepOnlyGraviolaSeries(seriesSets []storage.SeriesSet) []*domain.GraviolaSeries {
	graviolaSeries := make([]*domain.GraviolaSeries, 0, len(seriesSets)*2) //TODO: magic number, rework it
	for _, sSet := range seriesSets {
		gravSeriesSet, ok := sSet.(*domain.GraviolaSeriesSet)

		if ok && gravSeriesSet.Series != nil && len(gravSeriesSet.Series) > 0 {
			graviolaSeries = append(graviolaSeries, gravSeriesSet.Series...)
		}
	}

	return graviolaSeries
}

func mergeAnnotations(seriesSets []storage.SeriesSet) *annotations.Annotations {
	mergedAnnots := annotations.New()

	for _, seriesSet := range seriesSets {
		if seriesSet.Warnings() != nil {
			mergedAnnots.Merge(seriesSet.Warnings())
		}
	}

	return mergedAnnots
}
