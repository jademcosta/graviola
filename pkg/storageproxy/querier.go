package storageproxy

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/util/annotations"
)

type GraviolaQuerier struct {
	logg   *slog.Logger
	client *http.Client
}

func NewGraviolaQuerier(logg *slog.Logger) *GraviolaQuerier {
	return &GraviolaQuerier{
		logg:   logg,
		client: &http.Client{},
	}
}

// Querier
//
// Select returns a set of series that matches the given label matchers.
// Caller can specify if it requires returned series to be sorted. Prefer not requiring sorting for better performance.
// It allows passing hints that can help in optimising select, but it's up to implementation how this is used if used at all.
func (querier *GraviolaQuerier) Select(ctx context.Context, sortSeries bool, hints *storage.SelectHints, matchers ...*labels.Matcher) storage.SeriesSet {
	querier.logg.Debug("Select")

	for _, match := range matchers {
		querier.logg.Info("matchers internals", "match", fmt.Sprintf("%#v", match), "astring", match.String())
	}

	querier.logg.Info("hints internals", "matchers", fmt.Sprintf("%#v", hints))

	req, err := http.NewRequest(http.MethodGet, "http://localhost:9090/api/v1/query?query=up", nil)
	if err != nil {
		querier.logg.Error("error building the request", "error", err)
	}

	resp, err := querier.client.Do(req)
	if err != nil {
		querier.logg.Error("error performing the HTTP request", "error", err)
	}

	if !requestSuccessful(resp.StatusCode) {
		querier.logg.Error("request failed", "code", resp.StatusCode)
		return nil
	}

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		querier.logg.Error("reading body", "error", err)
	}
	defer resp.Body.Close()

	return nil
}

// {"msg":"matchers","match":"&labels.Matcher{Type:2, Name:\"nome\", Value:\"jade\", re:(*labels.FastRegexMatcher)(0xc0001794a0)}","astring":"nome=~\"jade\""}
// {"msg":"matchers","match":"&labels.Matcher{Type:0, Name:\"__name__\", Value:\"up\", re:(*labels.FastRegexMatcher)(nil)}","astring":"__name__=\"up\""}
// {"msg":"hints","matchers":"&storage.SelectHints{Start:1701638932075, End:1701639232075, Step:0, Func:\"\", Grouping:[]string(nil), By:false, Range:0, DisableTrimming:false}"}

// LabelQuerier
//
// Close releases the resources of the Querier.
func (querier *GraviolaQuerier) Close() error {
	querier.logg.Info("GraviolaQuerier: Close")
	return nil
}

// LabelQuerier
//
// LabelValues returns all potential values for a label name.
// It is not safe to use the strings beyond the lifetime of the querier.
// If matchers are specified the returned result set is reduced
// to label values of metrics matching the matchers.
func (querier *GraviolaQuerier) LabelValues(ctx context.Context, name string, matchers ...*labels.Matcher) ([]string, annotations.Annotations, error) {
	querier.logg.Info("GraviolaQuerier: LabelValues")
	return []string{"myownlabelval"}, map[string]error{}, nil
}

// LabelQuerier
//
// LabelNames returns all the unique label names present in the block in sorted order.
// If matchers are specified the returned result set is reduced
// to label names of metrics matching the matchers.
func (querier *GraviolaQuerier) LabelNames(ctx context.Context, matchers ...*labels.Matcher) ([]string, annotations.Annotations, error) {
	querier.logg.Info("GraviolaQuerier: LabelNames")
	return []string{"myownlabelnames"}, map[string]error{}, nil
}

func requestSuccessful(statusCode int) bool {
	return statusCode <= 299 && statusCode >= 200
}
