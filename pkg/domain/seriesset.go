package domain

import (
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/util/annotations"
)

type GraviolaSeriesSet struct {
	Series  []*GraviolaSeries
	current int
	Annots  annotations.Annotations
}

// SeriesSet
func (gSet *GraviolaSeriesSet) Next() bool {
	gSet.current++
	return gSet.current-1 < len(gSet.Series)
}

// SeriesSet
// At returns full series. Returned series should be iterable even after Next is called.
func (gSet *GraviolaSeriesSet) At() storage.Series {
	return gSet.Series[gSet.current-1]
}

// SeriesSet
// The error that iteration as failed with.
// When an error occurs, set cannot continue to iterate.
func (gSet *GraviolaSeriesSet) Err() error {
	return nil
}

// SeriesSet
// A collection of warnings for the whole set.
// Warnings could be return even iteration has not failed with error.
func (gSet *GraviolaSeriesSet) Warnings() annotations.Annotations {
	return gSet.Annots
}
