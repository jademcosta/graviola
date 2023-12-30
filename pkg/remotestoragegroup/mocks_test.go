package remotestoragegroup_test

import (
	"context"
	"sync"

	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/util/annotations"
)

type RemoteStorageMock struct {
	SeriesSet            *domain.GraviolaSeriesSet
	SelectFn             func(context.Context, bool, *storage.SelectHints, ...*labels.Matcher) storage.SeriesSet
	calledWithSortSeries []bool
	calledWithHints      []*storage.SelectHints
	calledWithMatchers   [][]*labels.Matcher
	closeCalled          int
	mu                   sync.Mutex
}

func (mock *RemoteStorageMock) Select(ctx context.Context, sortSeries bool, hints *storage.SelectHints, matchers ...*labels.Matcher) storage.SeriesSet {
	mock.mu.Lock()
	defer mock.mu.Unlock()

	mock.calledWithSortSeries = append(mock.calledWithSortSeries, sortSeries)
	mock.calledWithHints = append(mock.calledWithHints, hints)
	mock.calledWithMatchers = append(mock.calledWithMatchers, matchers)

	if mock.SelectFn != nil {
		return mock.SelectFn(ctx, sortSeries, hints, matchers...)
	}

	return mock.SeriesSet
}

func (mock *RemoteStorageMock) Close() error {
	mock.closeCalled++
	return nil
}

func (mock *RemoteStorageMock) LabelValues(ctx context.Context, name string, matchers ...*labels.Matcher) ([]string, annotations.Annotations, error) {

	lblVals := make([]string, 0)
	for _, serie := range mock.SeriesSet.Series {
		for _, val := range serie.Lbs.Map() {
			lblVals = append(lblVals, val)
		}
	}

	return lblVals, map[string]error{}, nil
}

func (mock *RemoteStorageMock) LabelNames(ctx context.Context, matchers ...*labels.Matcher) ([]string, annotations.Annotations, error) {
	lblNames := make([]string, 0)
	for _, serie := range mock.SeriesSet.Series {
		for _, val := range serie.Lbs.Map() {
			lblNames = append(lblNames, val)
		}
	}

	return lblNames, map[string]error{}, nil
}
