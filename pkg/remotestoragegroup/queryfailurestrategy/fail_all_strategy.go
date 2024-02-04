package queryfailurestrategy

import "github.com/prometheus/prometheus/storage"

type FailAllStrategy struct{}

// OnQueryFailureStrategy
func (fAllStrategy *FailAllStrategy) ForSeriesSet(sSets storage.SeriesSet) storage.SeriesSet {
	return sSets
}

// OnQueryFailureStrategy
func (fAllStrategy *FailAllStrategy) ForLabels(lbls []string, err error) ([]string, error) {
	return lbls, err
}
