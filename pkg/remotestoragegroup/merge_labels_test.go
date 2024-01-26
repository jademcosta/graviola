package remotestoragegroup_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jademcosta/graviola/internal/mocks"
	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/jademcosta/graviola/pkg/remotestoragegroup"
	"github.com/jademcosta/graviola/pkg/remotestoragegroup/mergestrategy"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/util/annotations"
	"github.com/stretchr/testify/assert"
)

func TestLabelValuesSuccessReturn(t *testing.T) {
	querier1 := &mocks.RemoteStorageMock{
		SeriesSet: &domain.GraviolaSeriesSet{
			Series: []*domain.GraviolaSeries{
				{Lbs: labels.FromStrings("__name__", "name1")},
				{Lbs: labels.FromStrings("__name__", "name2")},
				{Lbs: labels.FromStrings("__name__", "name3")},
				{Lbs: labels.FromStrings("__name__", "name3", "instance", "localhost:8080")},
			},
		},
	}

	querier2 := &mocks.RemoteStorageMock{
		SeriesSet: &domain.GraviolaSeriesSet{
			Series: []*domain.GraviolaSeries{
				{Lbs: labels.FromStrings("__name__", "name1")},
				{Lbs: labels.FromStrings("__name__", "name5", "instance", "localhost:9090")},
				{Lbs: labels.FromStrings("__name__", "name4")},
			},
		},
	}

	mergeStrategy := &mergestrategy.AlwaysMergeStrategy{}
	sut := remotestoragegroup.NewMergeQuerier([]storage.Querier{querier1, querier2}, mergeStrategy)

	vals, _, err := sut.LabelValues(context.Background(), "__name__")
	assert.NoError(t, err, "should return no error")
	assert.ElementsMatch(t, []string{"name1", "name2", "name3", "name4", "name5"}, vals,
		"should return correct label values")

	vals, _, err = sut.LabelValues(context.Background(), "instance")
	assert.NoError(t, err, "should return no error")
	assert.ElementsMatch(t, []string{"localhost:9090", "localhost:8080"}, vals,
		"should return correct label values")

	sut = remotestoragegroup.NewMergeQuerier([]storage.Querier{querier1}, mergeStrategy)

	vals, _, err = sut.LabelValues(context.Background(), "__name__")
	assert.NoError(t, err, "should return no error")
	assert.ElementsMatch(t, []string{"name1", "name2", "name3"}, vals, "should return correct label values")

	vals, _, err = sut.LabelValues(context.Background(), "non-existent")
	assert.NoError(t, err, "should return no error")
	assert.ElementsMatch(t, []string{}, vals,
		"should return empty when it does not have the label name equivalent")
}

func TestLabelValuesSendsTheQueryToAllRemotes(t *testing.T) {
	querier1 := &mocks.RemoteStorageMock{
		SeriesSet: &domain.GraviolaSeriesSet{
			Series: []*domain.GraviolaSeries{
				{Lbs: labels.FromStrings("__name__", "name1")},
			},
		},
	}

	querier2 := &mocks.RemoteStorageMock{
		SeriesSet: &domain.GraviolaSeriesSet{
			Series: []*domain.GraviolaSeries{
				{Lbs: labels.FromStrings("__name__", "name5", "instance", "localhost:9090")},
			},
		},
	}

	mergeStrategy := &mergestrategy.AlwaysMergeStrategy{}
	sut := remotestoragegroup.NewMergeQuerier([]storage.Querier{querier1, querier2}, mergeStrategy)

	ctx, cancelFn := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelFn()
	matchers := []*labels.Matcher{
		{Type: labels.MatchEqual, Name: "name1", Value: "val1"},
		{Type: labels.MatchEqual, Name: "name2", Value: "val2"},
	}
	_, _, _ = sut.LabelValues(ctx, "__name__", matchers...)
	assert.Equal(t, ctx, querier1.CalledWithContexts[0], "should have passed the context to remotes")
	assert.Equal(t, "__name__", querier1.CalledWithNames[0], "should have passed the name to remotes")
	assert.Equal(t, matchers, querier1.CalledWithMatchers[0], "should have passed the matchers to remotes")

	assert.Equal(t, ctx, querier2.CalledWithContexts[0], "should have passed the context to remotes")
	assert.Equal(t, "__name__", querier2.CalledWithNames[0], "should have passed the name to remotes")
	assert.Equal(t, matchers, querier2.CalledWithMatchers[0], "should have passed the matchers to remotes")
}

func TestLabelValuesErrorReturn(t *testing.T) {
	error1 := errors.New("error one")
	error2 := errors.New("error two")

	querier1 := &mocks.RemoteStorageMock{
		Error: error1,
	}

	querier2 := &mocks.RemoteStorageMock{
		Error: error2,
	}

	mergeStrategy := &mergestrategy.AlwaysMergeStrategy{}
	sut := remotestoragegroup.NewMergeQuerier([]storage.Querier{querier1, querier2}, mergeStrategy)

	expectedAnnots := annotations.New()
	expectedAnnots.Add(error1)
	expectedAnnots.Add(error2)

	_, annots, err := sut.LabelValues(context.Background(), "__name__")
	errorsReturned := extractErrors(annots)

	assert.ElementsMatch(t, []error{error1, error2}, errorsReturned, "should have added the errors as annotations")

	oneOfTheErrorsReturned := err == error1 || err == error2
	assert.True(t, oneOfTheErrorsReturned, "should return one of the errors")
}

func TestLabelValuesAnnotationsReturn(t *testing.T) {
	annots1 := annotations.New().Add(errors.New("some err1"))
	annots2 := annotations.New().Add(errors.New("some err2"))

	querier1 := &mocks.RemoteStorageMock{
		Annots: annots1,
	}

	querier2 := &mocks.RemoteStorageMock{
		Annots: annots2,
	}

	mergeStrategy := &mergestrategy.AlwaysMergeStrategy{}
	sut := remotestoragegroup.NewMergeQuerier([]storage.Querier{querier1, querier2}, mergeStrategy)

	expectedAnnots := annotations.New()
	expectedAnnots.Merge(annots1)
	expectedAnnots.Merge(annots2)

	_, annots, err := sut.LabelValues(context.Background(), "__name__")
	assert.NoError(t, err, "should return no error")

	assert.ElementsMatch(t, expectedAnnots.AsStrings("", 0), annots.AsStrings("", 0), "annotations should have been merged")
}

func TestLabelNamesSuccessReturn(t *testing.T) {
	querier1 := &mocks.RemoteStorageMock{
		SeriesSet: &domain.GraviolaSeriesSet{
			Series: []*domain.GraviolaSeries{
				{Lbs: labels.FromStrings("__name__", "name1")},
				{Lbs: labels.FromStrings("__name__", "name2")},
				{Lbs: labels.FromStrings("__name__", "name3")},
			},
		},
	}

	querier2 := &mocks.RemoteStorageMock{
		SeriesSet: &domain.GraviolaSeriesSet{
			Series: []*domain.GraviolaSeries{
				{Lbs: labels.FromStrings("__name__", "name1")},
				{Lbs: labels.FromStrings("__name__", "name5", "instance", "localhost:9090")},
				{Lbs: labels.FromStrings("somename", "aaaaa")},
			},
		},
	}

	mergeStrategy := &mergestrategy.AlwaysMergeStrategy{}
	sut := remotestoragegroup.NewMergeQuerier([]storage.Querier{querier1, querier2}, mergeStrategy)

	vals, _, err := sut.LabelNames(context.Background())
	assert.NoError(t, err, "should return no error")
	assert.ElementsMatch(t, []string{"__name__", "instance", "somename"}, vals,
		"should return correct label values")

	sut = remotestoragegroup.NewMergeQuerier([]storage.Querier{querier1}, mergeStrategy)

	vals, _, err = sut.LabelNames(context.Background())
	assert.NoError(t, err, "should return no error")
	assert.ElementsMatch(t, []string{"__name__"}, vals, "should return correct label values")
}

func TestLabelNamesSendsTheQueryToAllRemotes(t *testing.T) {
	querier1 := &mocks.RemoteStorageMock{
		SeriesSet: &domain.GraviolaSeriesSet{
			Series: []*domain.GraviolaSeries{
				{Lbs: labels.FromStrings("__name__", "name1")},
			},
		},
	}

	querier2 := &mocks.RemoteStorageMock{
		SeriesSet: &domain.GraviolaSeriesSet{
			Series: []*domain.GraviolaSeries{
				{Lbs: labels.FromStrings("__name__", "name5", "instance", "localhost:9090")},
			},
		},
	}

	mergeStrategy := &mergestrategy.AlwaysMergeStrategy{}
	sut := remotestoragegroup.NewMergeQuerier([]storage.Querier{querier1, querier2}, mergeStrategy)

	ctx, cancelFn := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelFn()
	matchers := []*labels.Matcher{
		{Type: labels.MatchEqual, Name: "name1", Value: "val1"},
		{Type: labels.MatchEqual, Name: "name2", Value: "val2"},
	}
	_, _, _ = sut.LabelNames(ctx, matchers...)
	assert.Equal(t, ctx, querier1.CalledWithContexts[0], "should have passed the context to remotes")
	assert.Equal(t, matchers, querier1.CalledWithMatchers[0], "should have passed the matchers to remotes")

	assert.Equal(t, ctx, querier2.CalledWithContexts[0], "should have passed the context to remotes")
	assert.Equal(t, matchers, querier2.CalledWithMatchers[0], "should have passed the matchers to remotes")
}

func TestLabelNamesErrorReturn(t *testing.T) {
	error1 := errors.New("error one")
	error2 := errors.New("error two")

	querier1 := &mocks.RemoteStorageMock{
		Error: error1,
	}

	querier2 := &mocks.RemoteStorageMock{
		Error: error2,
	}

	mergeStrategy := &mergestrategy.AlwaysMergeStrategy{}
	sut := remotestoragegroup.NewMergeQuerier([]storage.Querier{querier1, querier2}, mergeStrategy)

	expectedAnnots := annotations.New()
	expectedAnnots.Add(error1)
	expectedAnnots.Add(error2)

	_, annots, err := sut.LabelNames(context.Background())
	errorsReturned := extractErrors(annots)

	assert.ElementsMatch(t, []error{error1, error2}, errorsReturned, "should have added the errors as annotations")

	oneOfTheErrorsReturned := err == error1 || err == error2
	assert.True(t, oneOfTheErrorsReturned, "should return one of the errors")
}

func TestLabelNamesAnnotationsReturn(t *testing.T) {
	annots1 := annotations.New().Add(errors.New("some err1"))
	annots2 := annotations.New().Add(errors.New("some err2"))

	querier1 := &mocks.RemoteStorageMock{
		Annots: annots1,
	}

	querier2 := &mocks.RemoteStorageMock{
		Annots: annots2,
	}

	mergeStrategy := &mergestrategy.AlwaysMergeStrategy{}
	sut := remotestoragegroup.NewMergeQuerier([]storage.Querier{querier1, querier2}, mergeStrategy)

	expectedAnnots := annotations.New()
	expectedAnnots.Merge(annots1)
	expectedAnnots.Merge(annots2)

	_, annots, err := sut.LabelNames(context.Background())
	assert.NoError(t, err, "should return no error")

	assert.ElementsMatch(t, expectedAnnots.AsStrings("", 0), annots.AsStrings("", 0), "annotations should have been merged")
}

func extractErrors(annots annotations.Annotations) []error {
	errs := make([]error, 0)
	for _, err := range annots {
		errs = append(errs, err)
	}
	return errs
}
