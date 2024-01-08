package app_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/jademcosta/graviola/pkg/app"
	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestIntegrationAnswers422OnRemoteError(t *testing.T) {

	conf := config.GraviolaConfig{}
	err := yaml.Unmarshal([]byte(configOneGroupWithOneRemote), &conf)
	if err != nil {
		panic(err)
	}

	mock1 := NewMockRemote(make(map[string]mockRemoteRoute))

	mock1Srv := httptest.NewServer(mock1.mux)
	defer mock1Srv.Close()

	conf.StoragesConf.Groups[0].Servers[0].Address = mock1Srv.URL

	app := app.NewApp(conf)
	go func() {
		app.Start()
	}()

	defer func() {
		app.Stop()
	}()

	time.Sleep(200 * time.Millisecond)

	testCases := []struct {
		remoteFixedReturn int
	}{
		{http.StatusNotFound},
		{http.StatusBadRequest},
		{http.StatusForbidden},
		{http.StatusTooManyRequests},
		{http.StatusInternalServerError},
		{http.StatusServiceUnavailable},
		{http.StatusBadGateway},
	}

	for _, tc := range testCases {
		mock1.fixedStatusCodeAnswer = tc.remoteFixedReturn

		resp := doRequest("http://localhost:8091/api/v1/query", storage.SelectHints{},
			labels.MustNewMatcher(labels.MatchEqual, "lbl1", "val1"))

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		respJSON := make(map[string]string)
		err = json.Unmarshal(body, &respJSON)
		if err != nil {
			panic(err)
		}

		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, "HTTP status should be 422")
		assert.Equal(t, "error", respJSON["status"], "response JSON status should be error")
		assert.Equal(t, "execution", respJSON["errorType"], "response JSON errorType should be execution")
		assert.Equal(
			t,
			fmt.Sprintf("expanding series: server answered with non-succesful status code %d", tc.remoteFixedReturn),
			respJSON["error"],
			"response JSON error should inform the correct string",
		)
	}
}

func TestIntegrationQueryEngineUsesTheLookbackDelta(t *testing.T) {
	// The query engine has a Lookback delta number which tells how old a sample should be before
	// being considered stale. Due to that, it will convert instant queries into a range query

	conf := config.GraviolaConfig{}
	err := yaml.Unmarshal([]byte(configOneGroupWithOneRemote), &conf)
	if err != nil {
		panic(err)
	}

	mockRemote1 := NewMockRemote(make(map[string]mockRemoteRoute))

	mockRemote1Srv := httptest.NewServer(mockRemote1.mux)
	defer mockRemote1Srv.Close()

	conf.StoragesConf.Groups[0].Servers[0].Address = mockRemote1Srv.URL

	app := app.NewApp(conf)
	go func() {
		app.Start()
	}()

	defer func() {
		app.Stop()
	}()

	time.Sleep(200 * time.Millisecond)

	queryURL := "http://localhost:8091/api/v1/query"
	doRequest(queryURL, storage.SelectHints{}, labels.MustNewMatcher(labels.MatchEqual, "lbl1", "val1"))

	currentTime := time.Now()

	end := mustParseInt64(mockRemote1.calledWith[0].Form.Get("end"))
	start := mustParseInt64(mockRemote1.calledWith[0].Form.Get("start"))
	hundredMilliseconds := 100.0

	assert.InDelta(t, currentTime.Unix(), end, hundredMilliseconds, "should have sent the end parameter close to the current time")
	assert.InDelta(t, currentTime.Add(-5*time.Minute).Unix(), start, hundredMilliseconds,
		"should have sent the start parameter as the default value of '5 minutes ago'")
	assert.Equal(t, "30", mockRemote1.calledWith[0].Form.Get("step"), "the default step is NOT empty")
	assert.Equal(t, "{lbl1=\"val1\",}", mockRemote1.calledWith[0].Form.Get("query"), "the query should be present")

	queryURL = "http://localhost:8091/api/v1/query_range"
	doRequest(queryURL, storage.SelectHints{Start: 12145, End: 12595, Step: 11}, labels.MustNewMatcher(labels.MatchEqual, "lbl1", "val1"))

	end = mustParseInt64(mockRemote1.calledWith[1].Form.Get("end"))
	start = mustParseInt64(mockRemote1.calledWith[1].Form.Get("start"))
	step := mustParseInt64(mockRemote1.calledWith[1].Form.Get("step"))

	assert.Equal(t, int64(11845), start, "should have sent the start parameter provided, minus the 5 minutes stale data")
	assert.Equal(t, int64(12595), end, "should have sent the end parameter provided")
	assert.Equal(t, int64(11), step, "should have sent the step parameter provided")
	assert.Equal(t, "{lbl1=\"val1\",}", mockRemote1.calledWith[1].Form.Get("query"), "the query should be present")
}

func TestIntegrationSingleRemoteSuccess(t *testing.T) {
	conf := config.GraviolaConfig{}
	err := yaml.Unmarshal([]byte(configOneGroupWithOneRemote), &conf)
	panicOnError(err)

	mockRemote1 := NewMockRemote(make(map[string]mockRemoteRoute))

	mockRemote1Srv := httptest.NewServer(mockRemote1.mux)
	defer mockRemote1Srv.Close()

	conf.StoragesConf.Groups[0].Servers[0].Address = mockRemote1Srv.URL

	app := app.NewApp(conf)
	go func() {
		app.Start()
	}()

	defer func() {
		app.Stop()
	}()

	time.Sleep(200 * time.Millisecond)

	currentTime := time.Now()

	testCases := []struct {
		route        string
		hints        storage.SelectHints
		responseData mockRemoteRoute
	}{ //TODO: add more cases
		{"/api/v1/query_range",
			storage.SelectHints{
				Start: currentTime.Unix(), End: currentTime.Unix(), Step: 30,
			},
			mockRemoteRoute{
				status:     200,
				resultType: "matrix",
				series: &domain.GraviolaSeriesSet{
					Series: []*domain.GraviolaSeries{
						{
							Lbs:        labels.FromStrings("lbl111", "value111", "__name__", "my-metric"),
							Datapoints: []model.SamplePair{{Timestamp: model.Time(currentTime.Add(-time.Second).UnixMilli()), Value: 312.0}},
						},
						{
							Lbs:        labels.FromStrings("lbl111", "value222", "__name__", "my-metric"),
							Datapoints: []model.SamplePair{{Timestamp: model.Time(currentTime.Add(-1 * time.Minute).UnixMilli()), Value: 0.1}},
						},
						{
							Lbs:        labels.FromStrings("new-label", "new_value", "__name__", "my-metric"),
							Datapoints: []model.SamplePair{{Timestamp: model.Time(currentTime.Add(-2 * time.Minute).UnixMilli()), Value: 1.0}},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		mockRemote1.seriesSetMap[tc.route] = tc.responseData

		queryURL := "http://localhost:8091/api/v1/query_range"
		if tc.route != "" {
			queryURL = fmt.Sprintf("http://localhost:8091%s", tc.route)
		}

		resp := doRequest(
			queryURL, tc.hints, labels.MustNewMatcher(labels.MatchEqual, "lbl111", "value111"),
		)

		body, err := io.ReadAll(resp.Body)
		panicOnError(err)

		respJSON := make(map[string]*json.RawMessage)
		err = json.Unmarshal(body, &respJSON)
		panicOnError(err)

		dataJSON := make(map[string]*json.RawMessage)
		err = json.Unmarshal(*respJSON["data"], &dataJSON)
		panicOnError(err)

		assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP status should be 200")
		assert.Equal(t, "\"success\"", string(*respJSON["status"]), "response JSON status should be success")

		assert.Equal(t, "\"matrix\"", string(*dataJSON["resultType"]), "result type should be matrix")

		//Notice that it answers the "old" values as being at the current time
		resultTemp := fmt.Sprintf("[{\"metric\":{\"__name__\":\"my-metric\",\"lbl111\":\"value111\"},\"values\":[[%d,\"312\"]]},{\"metric\":{\"__name__\":\"my-metric\",\"lbl111\":\"value222\"},\"values\":[[%d,\"0.1\"]]},{\"metric\":{\"__name__\":\"my-metric\",\"new-label\":\"new_value\"},\"values\":[[%d,\"1\"]]}]",
			currentTime.Unix(), currentTime.Unix(), currentTime.Unix())
		assert.Equal(t, resultTemp, string(*dataJSON["result"]), "should have answered with the expected result inside the response")
	}
}

func mustParseInt64(i string) int64 {
	ret, err := strconv.ParseInt(i, 10, 64)
	if err != nil {
		panic(err)
	}
	return ret
}
