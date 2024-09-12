package queryfailurestrategy

import (
	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/prometheus/prometheus/storage"
)

type PartialResponseStrategy struct{}

// OnQueryFailureStrategy
func (fAllStrategy *PartialResponseStrategy) ForSeriesSet(sSets storage.SeriesSet) storage.SeriesSet {
	if sSets.Err() == nil {
		return sSets
	}

	parsedSet, ok := sSets.(*domain.GraviolaSeriesSet)
	if !ok {
		return sSets
	}

	if parsedSet.Series != nil && len(parsedSet.Series) == 0 {
		return sSets
	}

	if !isThereDataInAnySeries(parsedSet.Series) {
		return sSets
	}

	parsedSet.Erro = nil
	return parsedSet
}

// OnQueryFailureStrategy
func (fAllStrategy *PartialResponseStrategy) ForLabels(lbls []string, err error) ([]string, error) {
	if err == nil { //No error, keep everything
		return lbls, err
	}

	if len(lbls) == 0 { //This error needs to be reported in case it exists
		return lbls, err
	}

	return lbls, nil //Ignore errors, as there's a partial response
}

func isThereDataInAnySeries(series []*domain.GraviolaSeries) bool {
	for _, serie := range series {
		if len(serie.Datapoints) > 0 {
			return true
		}
	}

	return false
}
