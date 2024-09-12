package remotestorage_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/remotestorage"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/util/annotations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const labelValuesResponse = `{"status":"success","data":[{{VALUES}}]}`

func TestCorrectlyParsesLabelValuesSuccessResponse(t *testing.T) {

	testCases := []struct {
		labels []string
	}{
		{[]string{"__name__", "address", "endpoint", "event", "handler", "id", "instance", "interval", "job", "le", "listener_name", "machine", "major", "method", "minor", "mode", "mountpoint", "name", "nodename"}},
		{[]string{}},
		{[]string{"single_label_value"}},
		{[]string{"localhost:9090", "localhost:9100"}},
	}

	for _, tc := range testCases {
		mockRemote := MockRemote{
			mux: http.NewServeMux(),
		}

		mockRemote.mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			var content string
			if len(tc.labels) > 0 {
				content = "\"" + strings.Join(tc.labels, "\",\"") + "\""
			} else {
				content = ""
			}

			responseBody := strings.ReplaceAll(labelValuesResponse, "{{VALUES}}", content)

			_, err := w.Write([]byte(responseBody))
			panicOnError(err)
		})

		remoteSrv := httptest.NewServer(mockRemote.mux)

		sut := remotestorage.NewRemoteStorage(logg, config.RemoteConfig{Name: "test", Address: remoteSrv.URL}, func() time.Time { return frozenTime })
		result, annotations, err := sut.LabelValues(context.Background(), "any-name")
		require.NoError(t, err, "should have returned no error")
		assert.Empty(t, annotations.AsErrors(), "should have no annotations")

		assert.Len(t, result, len(tc.labels), "should have returned all the label names")
		assert.ElementsMatch(t, tc.labels, result, "elements should match")

		remoteSrv.Close()
	}
}

func TestLabelValuesKnowsHowToDealWithRemoteErrors(t *testing.T) {

	testCases := []struct {
		response       string
		responseStatus int
	}{
		{labelValuesResponse, 400},
		{labelValuesResponse, 500},
		{`{"status":"error","data":["__name__","address","endpoint"]}`, 200},
		{`{"status":"success","data":["__name__",}`, 200},
		{`{"st`, 200},
	}

	for _, tc := range testCases {
		mockRemote := MockRemote{
			mux: http.NewServeMux(),
		}

		mockRemote.mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(tc.responseStatus)
			_, err := w.Write([]byte(tc.response))
			panicOnError(err)
		})

		remoteSrv := httptest.NewServer(mockRemote.mux)

		sut := remotestorage.NewRemoteStorage(logg, config.RemoteConfig{Name: "test", Address: remoteSrv.URL}, func() time.Time { return frozenTime })
		_, _, err := sut.LabelNames(context.Background())
		require.Errorf(t, err, "should have returned no error when status is %d and response %s",
			tc.responseStatus, tc.response)

		remoteSrv.Close()
	}
}

func TestLabelValuesParametersAreSentToRemote(t *testing.T) {

	var calledWithParams string
	var calledWithLabelName string
	mux := chi.NewMux()

	mux.HandleFunc(fmt.Sprintf(remotestorage.DefaultLabelValuesPath, "{labelname}"), func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`{"status":"success","data":["hi"]}`))
		assert.NoError(t, err, "should return no error")

		w.WriteHeader(http.StatusOK)
		err = r.ParseForm()
		assert.NoError(t, err, "should return no error")
		calledWithParams = r.Form.Encode()
		calledWithLabelName = chi.URLParam(r, "labelname")
	})

	remoteSrv := httptest.NewServer(mux)
	defer remoteSrv.Close()

	matchers := []*labels.Matcher{
		{Type: labels.MatchEqual, Name: "label_filter_1", Value: "a value for here"},
		{Type: labels.MatchNotEqual, Name: "my_label", Value: "another value"},
	}

	sut := remotestorage.NewRemoteStorage(logg, config.RemoteConfig{Name: "test", Address: remoteSrv.URL}, func() time.Time { return frozenTime })
	_, _, err := sut.LabelValues(context.Background(), "some-random-name", matchers...)
	require.NoError(t, err, "should return no error")

	paramsResult, err := url.QueryUnescape(calledWithParams)
	require.NoError(t, err, "should return no error")
	assert.Equal(t, generateQueryParams(matchers), paramsResult, "query params should match")
	assert.Equal(t, "some-random-name", calledWithLabelName, "query params should match")
}

func TestLabelValuesWarningsAreTurnedIntoAnnotations(t *testing.T) {
	testCases := []struct {
		response string
		expected annotations.Annotations
	}{
		{
			`{"status":"success","data":["localhost:9090"],"warnings":["something went awfuly wrong"]}`,
			annotations.Annotations(map[string]error{"something went awfuly wrong": errors.New("something went awfuly wrong")}),
		},
		{
			`{"status":"success","data":["localhost:9090"],"warnings":["something went awfuly wrong", "agaaaain"]}`,
			annotations.Annotations(map[string]error{
				"something went awfuly wrong": errors.New("something went awfuly wrong"),
				"agaaaain":                    errors.New("agaaaain"),
			}),
		},
		{
			`{"status":"success","data":["localhost:9090"],"warnings":[]}`,
			annotations.Annotations(map[string]error{}),
		},
	}

	for _, tc := range testCases {
		mockRemote := MockRemote{
			mux: http.NewServeMux(),
		}

		mockRemote.mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(tc.response))
			panicOnError(err)
		})

		remoteSrv := httptest.NewServer(mockRemote.mux)

		sut := remotestorage.NewRemoteStorage(logg, config.RemoteConfig{Name: "test", Address: remoteSrv.URL}, func() time.Time { return frozenTime })
		_, annots, err := sut.LabelValues(context.Background(), "any-name")
		require.NoError(t, err, "should have returned NO error")
		assert.Equal(
			t,
			tc.expected,
			annots,
			"annotations should match",
		)

		remoteSrv.Close()
	}
}
