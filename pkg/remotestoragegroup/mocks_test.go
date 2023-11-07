package remotestoragegroup_test

import (
	"context"

	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/util/annotations"
)

type RemoteStorageMock struct {
	SeriesSet            *domain.GraviolaSeriesSet
	calledWithSortSeries []bool
	calledWithHints      []*storage.SelectHints
	calledWithMatchers   [][]*labels.Matcher
}

func (mock *RemoteStorageMock) Select(ctx context.Context, sortSeries bool, hints *storage.SelectHints, matchers ...*labels.Matcher) storage.SeriesSet {
	mock.calledWithSortSeries = append(mock.calledWithSortSeries, sortSeries)
	mock.calledWithHints = append(mock.calledWithHints, hints)
	mock.calledWithMatchers = append(mock.calledWithMatchers, matchers)

	return mock.SeriesSet
}

func (mock *RemoteStorageMock) Close() error {
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
