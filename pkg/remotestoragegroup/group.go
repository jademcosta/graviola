package remotestoragegroup

import (
	"context"
	"log/slog"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/util/annotations"
)

type OnQueryFailureStrategy interface {
	ForSeriesSet(storage.SeriesSet) storage.SeriesSet
	ForLabels([]string, error) ([]string, error)
}

type Group struct {
	Name              string
	remoteStorages    []storage.Querier
	remoteStoragesLen int
	onQueryFailure    OnQueryFailureStrategy
	logg              *slog.Logger
	seriesSetMerger   MergeStrategy
}

func NewGroup(logg *slog.Logger, name string, remoteStorages []storage.Querier,
	onQueryFailure OnQueryFailureStrategy, mergeStrategy MergeStrategy) *Group {

	return &Group{
		Name:              name,
		remoteStorages:    remoteStorages,
		remoteStoragesLen: len(remoteStorages),
		logg:              logg.With("name", name, "component", "group"),
		seriesSetMerger:   mergeStrategy,
		onQueryFailure:    onQueryFailure,
	}
}

// Querier
// Select returns a set of series that matches the given label matchers.
// Caller can specify if it requires returned series to be sorted. Prefer not requiring sorting for better performance.
// It allows passing hints that can help in optimising select, but it's up to implementation how this is used if used at all.
func (grp *Group) Select(ctx context.Context, sortSeries bool, hints *storage.SelectHints, matchers ...*labels.Matcher) storage.SeriesSet {
	mergeQuerier := NewMergeQuerier(grp.remoteStorages, grp.seriesSetMerger)

	response := mergeQuerier.Select(ctx, sortSeries, hints, matchers...)
	return grp.onQueryFailure.ForSeriesSet(response)
}

// LabelQuerier
// Close releases the resources of the Querier.
func (grp *Group) Close() error {
	mergeQuerier := NewMergeQuerier(grp.remoteStorages, grp.seriesSetMerger)

	response := mergeQuerier.Close()
	return response
}

// LabelQuerier
// LabelValues returns all potential values for a label name.
// It is not safe to use the strings beyond the lifetime of the querier.
// If matchers are specified the returned result set is reduced
// to label values of metrics matching the matchers.
func (grp *Group) LabelValues(ctx context.Context, name string, hints *storage.LabelHints, matchers ...*labels.Matcher) ([]string, annotations.Annotations, error) {
	mergeQuerier := NewMergeQuerier(grp.remoteStorages, grp.seriesSetMerger)

	vals, annots, err := mergeQuerier.LabelValues(ctx, name, hints, matchers...)
	vals, err = grp.onQueryFailure.ForLabels(vals, err)
	return vals, annots, err
}

// LabelQuerier
// LabelNames returns all the unique label names present in the block in sorted order.
// If matchers are specified the returned result set is reduced
// to label names of metrics matching the matchers.
func (grp *Group) LabelNames(ctx context.Context, hints *storage.LabelHints, matchers ...*labels.Matcher) ([]string, annotations.Annotations, error) {
	mergeQuerier := NewMergeQuerier(grp.remoteStorages, grp.seriesSetMerger)

	vals, annots, err := mergeQuerier.LabelNames(ctx, hints, matchers...)
	vals, err = grp.onQueryFailure.ForLabels(vals, err)

	return vals, annots, err
}
