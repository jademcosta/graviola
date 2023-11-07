package remotestoragegroup

import (
	"context"

	"github.com/jademcosta/graviola/pkg/graviolalog"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/util/annotations"
)

type Group struct {
	Name              string
	remoteStorages    []storage.Querier
	remoteStoragesLen int
	querier           storage.Querier
	// StrategyOnQueryFailure string //TODO: implement me
	// dedupInterval  *time.Duration
	logg *graviolalog.Logger
}

func NewGroup(logg *graviolalog.Logger, name string, remoteStorages []storage.Querier) *Group {
	return &Group{
		Name:              name,
		remoteStorages:    remoteStorages,
		remoteStoragesLen: len(remoteStorages),
		logg:              logg,
		querier:           storage.NewMergeQuerier(remoteStorages, nil, storage.ChainedSeriesMerge),
	}
}

// Querier
//
// Select returns a set of series that matches the given label matchers.
// Caller can specify if it requires returned series to be sorted. Prefer not requiring sorting for better performance.
// It allows passing hints that can help in optimising select, but it's up to implementation how this is used if used at all.
func (grp *Group) Select(ctx context.Context, sortSeries bool, hints *storage.SelectHints, matchers ...*labels.Matcher) storage.SeriesSet {
	// seriesSet := grp.querier.Select(sortSeries, hints, matchers...)
	// graviolaSeriesSet, ok := seriesSet.(*domain.GraviolaSeriesSet)
	// if !ok {
	// 	fmt.Println("Not ok")
	// 	fmt.Printf("series is Type %T\n", seriesSet)
	// 	return storage.NoopSeriesSet()
	// }

	// slices.SortFunc(graviolaSeriesSet.Series, func(a, b *domain.GraviolaSeries) int {
	// 	return labels.Compare(a.Labels(), b.Labels())
	// })
	// return graviolaSeriesSet
	return grp.querier.Select(ctx, sortSeries, hints, matchers...)
}

// LabelQuerier
//
// Close releases the resources of the Querier.
func (grp *Group) Close() error {
	return nil
}

// LabelQuerier
//
//	// LabelValues returns all potential values for a label name.
//	// It is not safe to use the strings beyond the lifetime of the querier.
//	// If matchers are specified the returned result set is reduced
//	// to label values of metrics matching the matchers.
func (grp *Group) LabelValues(name string, matchers ...*labels.Matcher) ([]string, annotations.Annotations, error) {
	grp.logg.Info("Group: LabelValues") //TODO: implement me
	return []string{"myownlabelval"}, map[string]error{}, nil
}

// LabelQuerier
//
//	// LabelNames returns all the unique label names present in the block in sorted order.
//	// If matchers are specified the returned result set is reduced
//	// to label names of metrics matching the matchers.
func (grp *Group) LabelNames(matchers ...*labels.Matcher) ([]string, annotations.Annotations, error) {
	grp.logg.Info("GraviolaQuerier: LabelNames") //TODO: implement me
	return []string{"myownlabelnames"}, map[string]error{}, nil
}
