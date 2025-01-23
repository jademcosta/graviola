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
	CalledWithNames      []string
	CalledWithContexts   []context.Context
	CloseCalled          int
	Error                error
	Annots               annotations.Annotations
	Mu                   sync.Mutex
}

func (mock *RemoteStorageMock) Select(
	ctx context.Context, sortSeries bool, hints *storage.SelectHints, matchers ...*labels.Matcher,
) storage.SeriesSet {
	mock.Mu.Lock()
	defer mock.Mu.Unlock()

	mock.CalledWithContexts = append(mock.CalledWithContexts, ctx)
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
	if mock.Error != nil {
		return mock.Error
	}
	return nil
}

func (mock *RemoteStorageMock) LabelValues(
	ctx context.Context, name string, _ *storage.LabelHints, matchers ...*labels.Matcher, //TODO: test hints
) ([]string, annotations.Annotations, error) {
	mock.Mu.Lock()
	defer mock.Mu.Unlock()

	annots := annotations.New()
	annots.Merge(mock.Annots)

	mock.CalledWithContexts = append(mock.CalledWithContexts, ctx)
	mock.CalledWithNames = append(mock.CalledWithNames, name)
	mock.CalledWithMatchers = append(mock.CalledWithMatchers, matchers)

	lblVals := make([]string, 0)

	if mock.SeriesSet != nil {
		for _, serie := range mock.SeriesSet.Series {
			for lblName, val := range serie.Lbs.Map() {
				if lblName == name {
					lblVals = append(lblVals, val)
				}
			}
		}
	}

	var err error
	if mock.Error != nil {
		err = mock.Error
	}

	return lblVals, *annots, err
}

func (mock *RemoteStorageMock) LabelNames(
	ctx context.Context, _ *storage.LabelHints, matchers ...*labels.Matcher, //TODO: test hints
) ([]string, annotations.Annotations, error) {
	mock.Mu.Lock()
	defer mock.Mu.Unlock()

	annots := annotations.New()
	annots.Merge(mock.Annots)

	mock.CalledWithContexts = append(mock.CalledWithContexts, ctx)
	mock.CalledWithMatchers = append(mock.CalledWithMatchers, matchers)

	lblNames := make([]string, 0)
	if mock.SeriesSet != nil {
		for _, serie := range mock.SeriesSet.Series {
			for name := range serie.Lbs.Map() {
				lblNames = append(lblNames, name)
			}
		}
	}

	var err error
	if mock.Error != nil {
		err = mock.Error
	}

	return lblNames, *annots, err
}
