package app_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jademcosta/graviola/pkg/app"
	"github.com/jademcosta/graviola/pkg/config"
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

	mock1 := NewMockRemote(make(map[string]http.HandlerFunc))

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
