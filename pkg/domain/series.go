package domain

import (
	"sort"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/tsdb/chunkenc"
)

// The representation of a time-series, with its labels and datapoints
// Implements the Prometheus Series interface.
type GraviolaSeries struct {
	Lbs        labels.Labels
	Datapoints []model.SamplePair
}

func (gSerie *GraviolaSeries) Labels() labels.Labels {
	return gSerie.Lbs.Copy()
}

func (gSerie *GraviolaSeries) Iterator(iter chunkenc.Iterator) chunkenc.Iterator {
	if gravIter, ok := iter.(*GraviolaIterator); ok {
		gravIter.reset(gSerie)
		return gravIter
	}
	return newGraviolaIterator(gSerie)
}

// Iterator based on concreteSeriesIterator
// at https://github.com/prometheus/prometheus/blob/main/storage/remote/codec.go#L387
type GraviolaIterator struct {
	cur        int
	curValType chunkenc.ValueType
	series     *GraviolaSeries
}

func newGraviolaIterator(series *GraviolaSeries) chunkenc.Iterator {
	return &GraviolaIterator{
		cur:        -1,
		curValType: chunkenc.ValNone,
		series:     series,
	}
}

// Next advances the iterator by one and returns the type of the value
// at the new position (or ValNone if the iterator is exhausted).
func (gravIter *GraviolaIterator) Next() chunkenc.ValueType {

	if gravIter.cur+1 < len(gravIter.series.Datapoints) {
		gravIter.cur++
		gravIter.curValType = chunkenc.ValFloat
		return gravIter.curValType
	}
	gravIter.curValType = chunkenc.ValNone
	return chunkenc.ValNone
}

// Seek advances the iterator forward to the first sample with a
// timestamp equal or greater than t. If the current sample found by a
// previous `Next` or `Seek` operation already has this property, Seek
// has no effect. If a sample has been found, Seek returns the type of
// its value. Otherwise, it returns ValNone, after which the iterator is
// exhausted.
//
//nolint:stdmethods
func (gravIter *GraviolaIterator) Seek(t int64) chunkenc.ValueType {
	if gravIter.cur == -1 {
		gravIter.cur = 0
	}

	if gravIter.cur >= len(gravIter.series.Datapoints) {
		return chunkenc.ValNone
	}

	// No-op check.
	if gravIter.curValType == chunkenc.ValFloat && int64(gravIter.series.Datapoints[gravIter.cur].Timestamp) >= t {
		return gravIter.curValType
	}

	gravIter.curValType = chunkenc.ValNone

	// Binary search between current position and end for both float and histograms samples.
	gravIter.cur += sort.Search(len(gravIter.series.Datapoints)-gravIter.cur, func(n int) bool {
		return int64(gravIter.series.Datapoints[n+gravIter.cur].Timestamp) >= t
	})

	if gravIter.cur < len(gravIter.series.Datapoints) {
		gravIter.curValType = chunkenc.ValFloat
	}

	return gravIter.curValType
}

// At returns the current timestamp/value pair if the value is a float.
// Before the iterator has advanced, the behaviour is unspecified.
func (gravIter *GraviolaIterator) At() (int64, float64) {
	if gravIter.curValType != chunkenc.ValFloat {
		panic("iterator is not on a float sample")
	}

	datapoint := gravIter.series.Datapoints[gravIter.cur]
	return int64(datapoint.Timestamp), float64(datapoint.Value)
}

// AtHistogram returns the current timestamp/value pair if the value is
// a histogram with integer counts. Before the iterator has advanced,
// the behaviour is unspecified.
func (gravIter *GraviolaIterator) AtHistogram(_ *histogram.Histogram) (int64, *histogram.Histogram) {
	panic("native histograms is not supported yet") //TODO: add support
}

// AtFloatHistogram returns the current timestamp/value pair if the
// value is a histogram with floating-point counts. It also works if the
// value is a histogram with integer counts, in which case a
// FloatHistogram copy of the histogram is returned. Before the iterator
// has advanced, the behaviour is unspecified.
func (gravIter *GraviolaIterator) AtFloatHistogram(_ *histogram.FloatHistogram) (int64, *histogram.FloatHistogram) {
	panic("implement-me") //FIXME: implement this
}

// AtT returns the current timestamp.
// Before the iterator has advanced, the behaviour is unspecified.
func (gravIter *GraviolaIterator) AtT() int64 {
	return int64(gravIter.series.Datapoints[gravIter.cur].Timestamp)
}

// Err returns the current error. It should be used only after the
// iterator is exhausted, i.e. `Next` or `Seek` have returned ValNone.
func (gravIter *GraviolaIterator) Err() error {
	return nil
}

func (gravIter *GraviolaIterator) reset(series *GraviolaSeries) {
	gravIter.cur = -1
	gravIter.curValType = chunkenc.ValNone
	gravIter.series = series
}
