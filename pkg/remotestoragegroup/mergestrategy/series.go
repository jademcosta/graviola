package mergestrategy

import (
	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/prometheus/prometheus/storage"
)

func keepOnlyGraviolaSeries(seriesSets []storage.SeriesSet) []*domain.GraviolaSeries {
	graviolaSeries := make([]*domain.GraviolaSeries, 0, len(seriesSets)*2) //TODO: magic number, rework it
	for _, sSet := range seriesSets {
		gravSeriesSet, ok := sSet.(*domain.GraviolaSeriesSet)

		if ok {
			graviolaSeries = append(graviolaSeries, gravSeriesSet.Series...)
		}
	}

	return graviolaSeries
}
