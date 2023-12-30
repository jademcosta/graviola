package mocks

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
	CalledWithSortSeries []bool
	CalledWithHints      []*storage.SelectHints
	CalledWithMatchers   [][]*labels.Matcher
	CloseCalled          int
	Mu                   sync.Mutex
}

func (mock *RemoteStorageMock) Select(ctx context.Context, sortSeries bool, hints *storage.SelectHints, matchers ...*labels.Matcher) storage.SeriesSet {
	mock.Mu.Lock()
	defer mock.Mu.Unlock()

	mock.CalledWithSortSeries = append(mock.CalledWithSortSeries, sortSeries)
	mock.CalledWithHints = append(mock.CalledWithHints, hints)
	mock.CalledWithMatchers = append(mock.CalledWithMatchers, matchers)

	if mock.SelectFn != nil {
		return mock.SelectFn(ctx, sortSeries, hints, matchers...)
	}

	return mock.SeriesSet
}

func (mock *RemoteStorageMock) Close() error {
	mock.CloseCalled++
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
