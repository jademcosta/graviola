package remotestoragegroup

import (
	"context"
	"sync"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/util/annotations"
)

type labelResponse struct {
	values []string
	annots annotations.Annotations
	err    error
}

type MergeStrategy interface {
	Merge([]storage.SeriesSet) storage.SeriesSet
}

type mergeQuerier struct {
	queriers        []storage.Querier
	seriesSetMerger MergeStrategy
}

func NewMergeQuerier(queriers []storage.Querier, seriesSetMerger MergeStrategy) *mergeQuerier {
	if seriesSetMerger == nil {
		panic("the merge strategy cannot be nil when creating a MergeQuerier")
	}

	return &mergeQuerier{
		queriers:        queriers,
		seriesSetMerger: seriesSetMerger,
	}
}

// Querier
// Select returns a set of series that matches the given label matchers.
// Caller can specify if it requires returned series to be sorted. Prefer not requiring sorting for better performance.
// It allows passing hints that can help in optimising select, but it's up to implementation how this is used if used at all.
func (mq *mergeQuerier) Select(ctx context.Context, sortSeries bool, hints *storage.SelectHints, matchers ...*labels.Matcher) storage.SeriesSet {
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
	response := mq.seriesSetMerger.Merge(seriesSets)
	return response
}

// LabelQuerier
// Close releases the resources of the Querier.
func (mq *mergeQuerier) Close() error {
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
func (mq *mergeQuerier) LabelValues(ctx context.Context, name string, matchers ...*labels.Matcher) ([]string, annotations.Annotations, error) {

	if len(mq.queriers) == 0 {
		return []string{}, map[string]error{}, nil
	}

	if len(mq.queriers) == 1 {
		return mq.queriers[0].LabelValues(ctx, name, matchers...)
	}

	labelValuesSet := make(map[string]struct{})
	errs := make([]error, 0)
	annots := annotations.New()

	var wg sync.WaitGroup
	valuesChan := make(chan *labelResponse)

	wg.Add(len(mq.queriers))
	for _, querier := range mq.queriers {
		go func(qr storage.Querier) {

			defer wg.Done()
			values, annotationsResponse, err := qr.LabelValues(ctx, name, matchers...)

			valuesChan <- &labelResponse{
				values: values,
				annots: annotationsResponse,
				err:    err,
			}
		}(querier)
	}

	go func() {
		wg.Wait()
		close(valuesChan)
	}()

	for lblResp := range valuesChan {
		annots.Merge(lblResp.annots)
		errs = append(errs, lblResp.err)
		for _, val := range lblResp.values {
			labelValuesSet[val] = struct{}{}
		}
	}

	valuesDeduped := make([]string, 0, len(labelValuesSet))
	for val := range labelValuesSet {
		valuesDeduped = append(valuesDeduped, val)
	}

	var err error = nil
	if len(errs) > 0 {
		err = errs[0]
	}

	//TODO: check if context is cancelled
	return valuesDeduped, *annots, err
}

// LabelQuerier
// LabelNames returns all the unique label names present in the block in sorted order.
// If matchers are specified the returned result set is reduced
// to label names of metrics matching the matchers.
func (mq *mergeQuerier) LabelNames(ctx context.Context, matchers ...*labels.Matcher) ([]string, annotations.Annotations, error) {
	//TODO: implement me
	return []string{"myownlabelnames"}, map[string]error{}, nil
}
