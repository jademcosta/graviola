package remotestorage

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/jademcosta/graviola/pkg/graviolalog"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/util/annotations"
	api_v1 "github.com/prometheus/prometheus/web/api/v1"
)

const DefaultInstantQueryPath = "/api/v1/query"
const DefaultRangeQueryPath = "/api/v1/query_range"

type RemoteStorage struct {
	logg   *graviolalog.Logger
	URLs   map[string]string
	client *http.Client
	now    func() time.Time
}

func NewRemoteStorage(logg *graviolalog.Logger, conf config.RemoteConfig, now func() time.Time) *RemoteStorage {
	//TODO: add WITH on the logger
	return &RemoteStorage{
		logg:   logg,
		URLs:   generateURLs(conf, logg),
		client: &http.Client{}, //TODO: allow to config this
		now:    now,
	}
}

// Querier
//
// Select returns a set of series that matches the given label matchers.
// Caller can specify if it requires returned series to be sorted. Prefer not requiring sorting for better performance.
// It allows passing hints that can help in optimising select, but it's up to implementation how this is used if used at all.
func (rStorage *RemoteStorage) Select(sortSeries bool, hints *storage.SelectHints, matchers ...*labels.Matcher) storage.SeriesSet {

	// now := rStorage.now().Unix()

	promQLQuery, err := ToPromQLQuery(matchers)
	if err != nil {
		rStorage.logg.Error("error creating query params", "error", err)
		return storage.NoopSeriesSet()
	}

	params := url.Values{}
	params.Set("query", *promQLQuery)

	var urlForQuery string
	if hints.End == 0 || hints.Start == 0 {
		urlForQuery = rStorage.URLs["instant_query"]
	} else {
		params.Set("start", fmt.Sprintf("%d", hints.Start))
		params.Set("end", fmt.Sprintf("%d", hints.End))
		if hints.Step != 0 {
			params.Set("step", fmt.Sprintf("%d", hints.Step))
		}

		urlForQuery = rStorage.URLs["range_query"]
	}

	req, err := http.NewRequest(http.MethodPost, urlForQuery, strings.NewReader(params.Encode()))
	if err != nil {
		rStorage.logg.Error("error creating request", "error", err)
		return storage.NoopSeriesSet()
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := rStorage.client.Do(req)
	if err != nil {
		rStorage.logg.Error("error making request", "error", err)
		return storage.NoopSeriesSet()
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		rStorage.logg.Error("error reading body", "error", err)
		return storage.NoopSeriesSet()
	}

	if !responseSuccessful(resp.StatusCode) {
		rStorage.logg.Error("server answered with non-succesful status code", "code", resp.StatusCode)
		return storage.NoopSeriesSet()
	}

	responseFromServer, err := parseResponse(data)
	if err != nil {
		rStorage.logg.Error("unable to parse target server response", "error", err)
		return storage.NoopSeriesSet()
	}

	if responseFromServer.Status == prometheusStatusError {
		rStorage.logg.Error("parsed response informed failure", "error", responseFromServer.Error)
		return storage.NoopSeriesSet()
	}
	//TODO: usar json.RawMessage https://mariadesouza.com/2017/09/07/custom-unmarshal-json-in-golang/

	reencodedData, err := json.Marshal(responseFromServer.Data)
	if err != nil {
		rStorage.logg.Error("error when reencoding data part of response", "error", err)
		return storage.NoopSeriesSet()
	}

	responseTSData, err := rStorage.parseTimeSeriesData(reencodedData, sortSeries)
	if err != nil {
		rStorage.logg.Error("unable to parse time-series data", "error", err)
		return storage.NoopSeriesSet()
	}

	return responseTSData
	//TODO: in case of warnings, how to proceed?
}

// LabelQuerier
//
// Close releases the resources of the Querier.
func (rStorage *RemoteStorage) Close() error {
	rStorage.logg.Info("Group: Close")
	return nil
}

// LabelQuerier
//
//	// LabelValues returns all potential values for a label name.
//	// It is not safe to use the strings beyond the lifetime of the querier.
//	// If matchers are specified the returned result set is reduced
//	// to label values of metrics matching the matchers.
func (rStorage *RemoteStorage) LabelValues(name string, matchers ...*labels.Matcher) ([]string, annotations.Annotations, error) {
	rStorage.logg.Info("Group: LabelValues")
	return []string{"myownlabelval"}, map[string]error{}, nil
}

// LabelQuerier
//
//	// LabelNames returns all the unique label names present in the block in sorted order.
//	// If matchers are specified the returned result set is reduced
//	// to label names of metrics matching the matchers.
func (rStorage *RemoteStorage) LabelNames(matchers ...*labels.Matcher) ([]string, annotations.Annotations, error) {
	rStorage.logg.Info("GraviolaQuerier: LabelNames")
	return []string{"myownlabelnames"}, map[string]error{}, nil
}

func (rStorage *RemoteStorage) parseTimeSeriesData(data []byte, sorted bool) (storage.SeriesSet, error) {

	var resultMap map[string]*json.RawMessage
	err := json.Unmarshal(data, &resultMap)
	if err != nil {
		return storage.NoopSeriesSet(), fmt.Errorf("decoding data: %w", err)
	}

	var resultType string
	err = json.Unmarshal(*resultMap["resultType"], &resultType)
	if err != nil {
		return storage.NoopSeriesSet(), fmt.Errorf("decoding resultType %v: %w", *resultMap["resultType"], err)
	}

	resultValue, ok := resultMap["result"]
	if !ok {
		return storage.NoopSeriesSet(), fmt.Errorf("no result value")
	}

	switch parser.ValueType(resultType) {
	case parser.ValueTypeNone:
		return storage.NoopSeriesSet(), nil
	case parser.ValueTypeString:
		rStorage.logg.Error("parsing string result type is not supported yet")
		return storage.NoopSeriesSet(), nil
	case parser.ValueTypeScalar:
		rStorage.logg.Error("parsing scalar result type is not supported yet")
		return storage.NoopSeriesSet(), nil
	case parser.ValueTypeVector:
		return parseResultTypeVector(*resultValue, sorted)
	case parser.ValueTypeMatrix:
		return parseResultTypeMatrix(*resultValue, sorted)
	}

	return storage.NoopSeriesSet(), fmt.Errorf("invalid result type %s", resultType)
}

func generateURLs(conf config.RemoteConfig, logg *graviolalog.Logger) map[string]string {
	result := make(map[string]string)

	base := conf.Address
	if conf.PathPrefix != "" {
		base = urlJoin(base, conf.PathPrefix)
	}

	result["instant_query"] = urlJoin(base, DefaultInstantQueryPath)
	result["range_query"] = urlJoin(base, DefaultRangeQueryPath)

	return result
}

func urlJoin(base string, path string) string {
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(path, "/")
}

func responseSuccessful(statusCode int) bool {
	return statusCode >= 200 && statusCode <= 299
}

func parseResponse(data []byte) (*api_v1.Response, error) {
	resp := &api_v1.Response{}
	err := json.Unmarshal(data, resp)
	return resp, err
}

func parseResultTypeVector(data []byte, sorted bool) (storage.SeriesSet, error) {
	var metrics model.Vector
	err := json.Unmarshal(data, &metrics)
	if err != nil {
		return storage.NoopSeriesSet(), fmt.Errorf("error parsing result: %w", err)
	}

	lblBuilder := labels.NewScratchBuilder(8) //TODO: magic number
	series := make([]*domain.GraviolaSeries, 0, len(metrics))

	for _, sample := range metrics {
		sample := sample
		for lblName, lblVal := range sample.Metric {
			lblBuilder.Add(string(lblName), string(lblVal))
		}

		lblBuilder.Sort()
		series = append(series, &domain.GraviolaSeries{
			Lbs:        lblBuilder.Labels(),
			Datapoints: []model.SamplePair{{Timestamp: sample.Timestamp, Value: sample.Value}},
		})
		lblBuilder.Reset()
	}

	if sorted {
		slices.SortFunc(series, func(a, b *domain.GraviolaSeries) int {
			return labels.Compare(a.Labels(), b.Labels())
		})
	}

	return &domain.GraviolaSeriesSet{
		Series: series,
	}, nil
}

func parseResultTypeMatrix(data []byte, sorted bool) (storage.SeriesSet, error) {
	var metrics model.Matrix
	err := json.Unmarshal(data, &metrics)
	if err != nil {
		return storage.NoopSeriesSet(), fmt.Errorf("error parsing result: %w", err)
	}

	lblBuilder := labels.NewScratchBuilder(8) //TODO: magic number
	series := make([]*domain.GraviolaSeries, 0, len(metrics))

	for _, sample := range metrics {
		copySample := sample
		for lblName, lblVal := range copySample.Metric { //TODO: is it possible to cast this to a map[string]string ?
			lblBuilder.Add(string(lblName), string(lblVal))
		}

		pairs := make([]model.SamplePair, 0, len(copySample.Values))
		for _, pair := range copySample.Values {
			pair := pair
			pairs = append(pairs, model.SamplePair{Timestamp: pair.Timestamp, Value: pair.Value})
		}

		lblBuilder.Sort()
		series = append(series, &domain.GraviolaSeries{
			Lbs:        lblBuilder.Labels(),
			Datapoints: pairs,
		})
		lblBuilder.Reset()
	}

	if sorted && len(series) > 1 {
		slices.SortFunc(series, func(a, b *domain.GraviolaSeries) int {
			return labels.Compare(a.Labels(), b.Labels())
		})
	}

	return &domain.GraviolaSeriesSet{
		Series: series,
	}, nil
}

func ToPromQLQuery(matchers []*labels.Matcher) (*string, error) {
	var query strings.Builder

	_, err := query.WriteString("{")
	if err != nil {
		return nil, err
	}

	for _, matcher := range matchers {
		_, err := query.WriteString(matcher.String())
		if err != nil {
			return nil, err
		}

		_, err = query.WriteString(",")
		if err != nil {
			return nil, err
		}
	}

	_, err = query.WriteString("}")
	if err != nil {
		return nil, err
	}

	result := query.String()
	return &result, nil
}
