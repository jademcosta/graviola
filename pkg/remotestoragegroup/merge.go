package remotestoragegroup

import (
	"context"
	"sync"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/util/annotations"
)

type MergeStrategy interface {
	Merge([]storage.SeriesSet) storage.SeriesSet
}

type MergeQuerier struct {
	queriers        []storage.Querier
	seriesSetMerger MergeStrategy
}

func NewMergeQuerier(queriers []storage.Querier, seriesSetMerger MergeStrategy) *MergeQuerier {
	if seriesSetMerger == nil {
		panic("the merge strategy cannot be nil when creating a MergeQuerier")
	}

	return &MergeQuerier{
		queriers:        queriers,
		seriesSetMerger: seriesSetMerger,
	}
}

// Querier
// Select returns a set of series that matches the given label matchers.
// Caller can specify if it requires returned series to be sorted. Prefer not requiring sorting for better performance.
// It allows passing hints that can help in optimising select, but it's up to implementation how this is used if used at all.
func (mq *MergeQuerier) Select(ctx context.Context, sortSeries bool, hints *storage.SelectHints, matchers ...*labels.Matcher) storage.SeriesSet {
	if len(mq.queriers) == 0 {
		return storage.NoopSeriesSet()
	}

	if len(mq.queriers) == 1 {
		return mq.queriers[0].Select(ctx, sortSeries, hints, matchers...)
	}

	seriesSets := make([]storage.SeriesSet, 0, len(mq.queriers))

	var wg sync.WaitGroup
	seriesSetChan := make(chan storage.SeriesSet)

	for _, querier := range mq.queriers {
		wg.Add(1)
		go func(qr storage.Querier) {
			defer wg.Done()

			seriesSetChan <- qr.Select(ctx, true, hints, matchers...)
		}(querier)
	}

	go func() {
		wg.Wait()
		close(seriesSetChan)
	}()

	for r := range seriesSetChan {
		seriesSets = append(seriesSets, r)
	}

	//TODO: check if context is cancelled
	return mq.seriesSetMerger.Merge(seriesSets)
}

// LabelQuerier
// Close releases the resources of the Querier.
func (mq *MergeQuerier) Close() error {
	var err []error
	for _, querier := range mq.queriers {
		err1 := querier.Close()
		if err != nil {
			err = append(err, err1)
		}
	}
	if len(err) > 0 {
		return err[0]
	}

	return nil
}

// LabelQuerier
// LabelValues returns all potential values for a label name.
// It is not safe to use the strings beyond the lifetime of the querier.
// If matchers are specified the returned result set is reduced
// to label values of metrics matching the matchers.
func (mq *MergeQuerier) LabelValues(ctx context.Context, name string, matchers ...*labels.Matcher) ([]string, annotations.Annotations, error) {
	//TODO: implement me
	return []string{"myownlabelval"}, map[string]error{}, nil
}

// LabelQuerier
// LabelNames returns all the unique label names present in the block in sorted order.
// If matchers are specified the returned result set is reduced
// to label names of metrics matching the matchers.
func (mq *MergeQuerier) LabelNames(ctx context.Context, matchers ...*labels.Matcher) ([]string, annotations.Annotations, error) {
	//TODO: implement me
	return []string{"myownlabelnames"}, map[string]error{}, nil
}
