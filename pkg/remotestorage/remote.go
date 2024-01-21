package remotestorage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/util/annotations"
	api_v1 "github.com/prometheus/prometheus/web/api/v1"
)

const DefaultLabelNamesPath = "/api/v1/labels"
const DefaultInstantQueryPath = "/api/v1/query"
const DefaultRangeQueryPath = "/api/v1/query_range"
const DefaultStep = 30 //30 seconds

type RemoteStorage struct {
	logg   *slog.Logger
	URLs   map[string]string
	client *http.Client
	now    func() time.Time
}

func NewRemoteStorage(logg *slog.Logger, conf config.RemoteConfig, now func() time.Time) *RemoteStorage {
	return &RemoteStorage{
		logg:   logg.With("name", conf.Name, "component", "remote"),
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
func (rStorage *RemoteStorage) Select(ctx context.Context, sortSeries bool, hints *storage.SelectHints, matchers ...*labels.Matcher) storage.SeriesSet {
	//TODO: use context

	promQLQuery, err := ToPromQLQuery(matchers)
	if err != nil {
		e := fmt.Errorf("error creating query params: %w", err)
		rStorage.logg.Error("param creation", "error", e)
		return &domain.GraviolaSeriesSet{
			Erro:   e,
			Annots: map[string]error{"remote_storage": e},
		}
	}

	params := url.Values{}
	params.Set("query", *promQLQuery)

	var urlForQuery string
	if (hints.End == hints.Start) && (hints.End == 0) {
		urlForQuery = rStorage.URLs["instant_query"]

	} else if hints.End == hints.Start {
		urlForQuery = rStorage.URLs["instant_query"]
		params.Set("time", fmt.Sprintf("%d", removeMillisFromUnixTimestamp(hints.Start)))
	} else {
		params.Set("start", fmt.Sprintf("%d", removeMillisFromUnixTimestamp(hints.Start)))
		params.Set("end", fmt.Sprintf("%d", removeMillisFromUnixTimestamp(hints.End)))

		if hints.Step == 0 { //TODO: allow a default step to be set by configs
			params.Set("step", fmt.Sprintf("%d", DefaultStep))
		} else {
			// The engine turn step into milliseconds, but the API accepts only seconds
			params.Set("step", fmt.Sprintf("%d", hints.Step/1000))
		}

		urlForQuery = rStorage.URLs["range_query"]
	}

	responseFromServer, err := rStorage.doRequest(urlForQuery, params.Encode())
	if err != nil {
		return &domain.GraviolaSeriesSet{
			Erro:   err,
			Annots: map[string]error{"remote_storage": err},
		}
	}

	if responseFromServer.Status == prometheusStatusError {
		e := fmt.Errorf("parsed response informed failure %s", responseFromServer.Error)
		rStorage.logg.Error("answer informed failure", "error", e)
		return &domain.GraviolaSeriesSet{
			Erro:   e,
			Annots: map[string]error{"remote_storage": e},
		}
	}

	//TODO: is it possible to skip the reencoding using *json.RawMessage or something like that?
	reencodedData, err := json.Marshal(responseFromServer.Data)
	if err != nil {
		e := fmt.Errorf("error when reencoding data part of response: %w", err)
		rStorage.logg.Error("marshal failure", "error", e)
		return &domain.GraviolaSeriesSet{
			Erro:   e,
			Annots: map[string]error{"remote_storage": e},
		}
	}

	responseTSData := rStorage.parseTimeSeriesData(reencodedData, sortSeries)
	if err != nil {
		e := fmt.Errorf("unable to parse time-series data: %w", err)
		rStorage.logg.Error("parsing time-series data", "error", e)
		return &domain.GraviolaSeriesSet{
			Erro:   e,
			Annots: map[string]error{"remote_storage": e},
		}
	}

	if len(responseFromServer.Warnings) > 0 {
		responseTSData.Annots = *annotations.New()
		responseTSData.Annots.Add(fmt.Errorf("warnings: %v", responseFromServer.Warnings))
	}

	return responseTSData
}

// LabelQuerier
//
// Close releases the resources of the Querier.
func (rStorage *RemoteStorage) Close() error {
	//TODO: cancel all requests
	return nil
}

// LabelQuerier
//
//	// LabelValues returns all potential values for a label name.
//	// It is not safe to use the strings beyond the lifetime of the querier.
//	// If matchers are specified the returned result set is reduced
//	// to label values of metrics matching the matchers.
func (rStorage *RemoteStorage) LabelValues(ctx context.Context, name string, matchers ...*labels.Matcher) ([]string, annotations.Annotations, error) {
	//TODO: implement me
	return []string{"myownlabelval"}, map[string]error{}, nil
}

// LabelQuerier
//
//	// LabelNames returns all the unique label names present in the block in sorted order.
//	// If matchers are specified the returned result set is reduced
//	// to label names of metrics matching the matchers.
func (rStorage *RemoteStorage) LabelNames(
	ctx context.Context,
	matchers ...*labels.Matcher,
) ([]string, annotations.Annotations, error) {

	params := make([]string, len(matchers))
	for idx, m := range matchers {
		params[idx] = "match[]=" + url.QueryEscape(m.String())
	}

	reqBody := strings.Join(params, "&")

	response, err := rStorage.doRequest(rStorage.URLs["label_names"], reqBody)
	if err != nil {
		return []string{}, map[string]error{"remote_storage": err}, err
	}

	if response.Status == prometheusStatusError {
		e := fmt.Errorf("parsed response informed failure: %s", response.Error)
		rStorage.logg.Error("answer informed failure", "error", e)
		return []string{}, map[string]error{"remote_storage": e}, e
	}

	names, err := rStorage.parseLabelStringSlice(response.Data)
	if err != nil {
		return []string{}, map[string]error{"remote_storage": err}, err
	}

	return names, map[string]error{}, nil
}

func (rStorage *RemoteStorage) doRequest(url string, payload string) (*api_v1.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		e := fmt.Errorf("error creating request: %w", err)
		rStorage.logg.Error("request creation", "error", e)
		return nil, e
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rStorage.logg.Debug("performing request", "url", req.URL.String(), "headers", req.Header,
		"body", payload, "method", req.Method)

	resp, err := rStorage.client.Do(req)
	if err != nil {
		e := fmt.Errorf("error making request: %w", err)
		rStorage.logg.Error("request making", "error", e)
		return nil, e
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		e := fmt.Errorf("error reading request body: %w", err)
		rStorage.logg.Error("request body reading", "error", e)
		return nil, e
	}

	rStorage.logg.Debug("remote response", "body", string(data), "headers", resp.Header)

	if !responseSuccessful(resp.StatusCode) {
		e := fmt.Errorf("server answered with non-succesful status code %d", resp.StatusCode)
		rStorage.logg.Error("non-successful status code", "error", e)

		return nil, e
	}

	responseFromServer, err := parseResponse(data)
	if err != nil {
		e := fmt.Errorf("error parsing response from remote: %w", err)
		rStorage.logg.Error("parsing response body", "error", e)
		return nil, e
	}

	return responseFromServer, nil
}

func (rStorage *RemoteStorage) parseLabelStringSlice(data interface{}) ([]string, error) {

	unparsed, err := json.Marshal(data)
	if err != nil {
		rStorage.logg.Error("reencoding data", "error", err)
		return nil, err
	}

	result := make([]string, 0)
	err = json.Unmarshal(unparsed, &result)
	if err != nil {
		rStorage.logg.Error("parsing data", "error", err)
		return nil, err
	}

	return result, nil
}

func (rStorage *RemoteStorage) parseTimeSeriesData(data []byte, sorted bool) *domain.GraviolaSeriesSet {

	var resultMap map[string]*json.RawMessage
	err := json.Unmarshal(data, &resultMap)
	if err != nil {
		e := fmt.Errorf("decoding data: %w", err)
		rStorage.logg.Error("decoding", "error", e)
		return &domain.GraviolaSeriesSet{
			Erro:   e,
			Annots: map[string]error{"remote_storage": e},
		}
	}

	var resultType string
	err = json.Unmarshal(*resultMap["resultType"], &resultType)
	if err != nil {
		e := fmt.Errorf("decoding resultType: %w", err)
		rStorage.logg.Error("getting result type", "error", e)
		return &domain.GraviolaSeriesSet{
			Erro:   e,
			Annots: map[string]error{"remote_storage": e},
		}
	}

	resultValue, ok := resultMap["result"]
	if !ok {
		e := fmt.Errorf("empty result")
		rStorage.logg.Error("reading result", "error", e)
		return &domain.GraviolaSeriesSet{
			Erro:   e,
			Annots: map[string]error{"remote_storage": e},
		}
	}

	switch parser.ValueType(resultType) {
	case parser.ValueTypeNone:
		e := fmt.Errorf("valueType is 'none'")
		rStorage.logg.Error("decoding value type", "error", e)

		return &domain.GraviolaSeriesSet{
			Erro:   e,
			Annots: map[string]error{"remote_storage": e},
		}
	case parser.ValueTypeString: //TODO: implement me
		e := fmt.Errorf("parsing 'string' result type is not supported yet")
		rStorage.logg.Error("decoding value type", "error", e)

		return &domain.GraviolaSeriesSet{
			Erro:   e,
			Annots: map[string]error{"remote_storage": e},
		}
	case parser.ValueTypeScalar: //TODO: implement me
		e := fmt.Errorf("parsing 'scalar' result type is not supported yet")
		rStorage.logg.Error("decoding value type", "error", e)

		return &domain.GraviolaSeriesSet{
			Erro:   e,
			Annots: map[string]error{"remote_storage": e},
		}
	case parser.ValueTypeVector:
		result, err := parseResultTypeVector(*resultValue, sorted)
		if err != nil {
			e := fmt.Errorf("error parsing vector result type: %w", err)
			rStorage.logg.Error("parsing vector type", "error", e)

			return &domain.GraviolaSeriesSet{
				Erro:   e,
				Annots: map[string]error{"remote_storage": e},
			}
		}
		return result
	case parser.ValueTypeMatrix:
		result, err := parseResultTypeMatrix(*resultValue, sorted)
		if err != nil {
			e := fmt.Errorf("error parsing vector result type: %w", err)
			rStorage.logg.Error("parsing vector type", "error", e)

			return &domain.GraviolaSeriesSet{
				Erro:   e,
				Annots: map[string]error{"remote_storage": e},
			}
		}
		return result
	}

	e := fmt.Errorf("invalid result type %s", resultType)
	rStorage.logg.Error("decoding value type", "error", e)
	return &domain.GraviolaSeriesSet{
		Erro:   e,
		Annots: map[string]error{"remote_storage": e},
	}
}

func generateURLs(conf config.RemoteConfig, logg *slog.Logger) map[string]string {
	result := make(map[string]string)

	base := conf.Address
	if conf.PathPrefix != "" {
		base = urlJoin(base, conf.PathPrefix)
	}

	result["instant_query"] = urlJoin(base, DefaultInstantQueryPath)
	result["range_query"] = urlJoin(base, DefaultRangeQueryPath)
	result["label_names"] = urlJoin(base, DefaultLabelNamesPath)

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

func parseResultTypeVector(data []byte, sorted bool) (*domain.GraviolaSeriesSet, error) {
	var metrics model.Vector
	err := json.Unmarshal(data, &metrics)
	if err != nil {
		return nil, fmt.Errorf("error parsing result: %w", err)
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

func parseResultTypeMatrix(data []byte, sorted bool) (*domain.GraviolaSeriesSet, error) {
	var metrics model.Matrix
	err := json.Unmarshal(data, &metrics)
	if err != nil {
		return nil, fmt.Errorf("error parsing result: %w", err)
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

// The timestamp passed down has milliseconds in it, the API wants it without millis
func removeMillisFromUnixTimestamp(timestampWithMillis int64) int64 {
	return time.UnixMilli(timestampWithMillis).Unix()
}
