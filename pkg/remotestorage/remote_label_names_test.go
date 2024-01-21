package remotestorage_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/remotestorage"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/util/annotations"
	"github.com/stretchr/testify/assert"
)

const labelNamesResponse = `{"status":"success","data":[{{LABELS}}]}`

func TestCorrectlyParsesTheRemoteSuccessfulResponse(t *testing.T) {

	testCases := []struct {
		labels []string
	}{
		{[]string{"__name__", "address", "endpoint", "event", "handler", "id", "instance", "interval", "job", "le", "listener_name", "machine", "major", "method", "minor", "mode", "mountpoint", "name", "nodename"}},
		{[]string{}},
		{[]string{"single_label_name"}},
	}

	for _, tc := range testCases {
		mockRemote := MockRemote{
			mux: http.NewServeMux(),
		}

		mockRemote.mux.HandleFunc(remotestorage.DefaultLabelNamesPath, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			var content string
			if len(tc.labels) > 0 {
				content = "\"" + strings.Join(tc.labels, "\",\"") + "\""
			} else {
				content = ""
			}

			responseBody := strings.ReplaceAll(labelNamesResponse, "{{LABELS}}", content)

			_, err := w.Write([]byte(responseBody))
			panicOnError(err)
		})

		remoteSrv := httptest.NewServer(mockRemote.mux)

		sut := remotestorage.NewRemoteStorage(logg, config.RemoteConfig{Name: "test", Address: remoteSrv.URL}, func() time.Time { return frozenTime })
		result, annotations, err := sut.LabelNames(context.Background())
		assert.NoError(t, err, "should have returned no error")
		assert.Len(t, annotations.AsErrors(), 0, "should have no annotations")

		assert.Len(t, result, len(tc.labels), "should have returned all the label names")
		assert.ElementsMatch(t, tc.labels, result, "elements should match")

		remoteSrv.Close()
	}
}

func TestKnowsHowToDealWithRemoteErrors(t *testing.T) {

	testCases := []struct {
		response       string
		responseStatus int
	}{
		{labelNamesResponse, 400},
		{labelNamesResponse, 500},
		{`{"status":"error","data":["__name__","address","endpoint"]}`, 200},
		{`{"status":"success","data":["__name__",}`, 200},
		{`{"st`, 200},
	}

	for _, tc := range testCases {
		mockRemote := MockRemote{
			mux: http.NewServeMux(),
		}

		mockRemote.mux.HandleFunc(remotestorage.DefaultLabelNamesPath, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.responseStatus)
			_, err := w.Write([]byte(tc.response))
			panicOnError(err)
		})

		remoteSrv := httptest.NewServer(mockRemote.mux)

		sut := remotestorage.NewRemoteStorage(logg, config.RemoteConfig{Name: "test", Address: remoteSrv.URL}, func() time.Time { return frozenTime })
		_, _, err := sut.LabelNames(context.Background())
		assert.Errorf(t, err, "should have returned no error when status is %d and response %s",
			tc.responseStatus, tc.response)

		remoteSrv.Close()
	}
}

func TestParametersAreSentToRemote(t *testing.T) {

	var calledWith string
	mockRemote := MockRemote{
		mux: http.NewServeMux(),
	}

	mockRemote.mux.HandleFunc(remotestorage.DefaultLabelNamesPath, func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`{"status":"success","data":["hi"]}`))
		assert.NoError(t, err, "should return no error")

		w.WriteHeader(http.StatusOK)
		err = r.ParseForm()
		assert.NoError(t, err, "should return no error")
		calledWith = r.Form.Encode()
	})

	remoteSrv := httptest.NewServer(mockRemote.mux)
	defer remoteSrv.Close()

	matchers := []*labels.Matcher{
		{Type: labels.MatchEqual, Name: "label_filter_1", Value: "a value for here"},
		{Type: labels.MatchNotEqual, Name: "my_label", Value: "another value"},
	}

	sut := remotestorage.NewRemoteStorage(logg, config.RemoteConfig{Name: "test", Address: remoteSrv.URL}, func() time.Time { return frozenTime })
	_, _, err := sut.LabelNames(context.Background(), matchers...)
	assert.NoError(t, err, "should return no error")

	result, err := url.QueryUnescape(calledWith)
	assert.NoError(t, err, "should return no error")
	assert.Equal(t, generateQueryParams(matchers), result, "query params should match")
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func generateQueryParams(matchers []*labels.Matcher) string {
	builder := strings.Builder{}

	for _, matcher := range matchers {
		_, err := builder.Write([]byte("match[]="))
		panicOnError(err)
		_, err = builder.Write([]byte(matcher.String()))
		panicOnError(err)
		_, err = builder.Write([]byte("&"))
		panicOnError(err)
	}

	return strings.TrimRight(builder.String(), "&")
}

func TestWarningsAreTurnedIntoAnnotations(t *testing.T) {
	testCases := []struct {
		response string
		expected annotations.Annotations
	}{
		{
			`{"status":"success","data":["__name__"],"warnings":["something went awfuly wrong"]}`,
			annotations.Annotations(map[string]error{"something went awfuly wrong": errors.New("something went awfuly wrong")}),
		},
		{
			`{"status":"success","data":["__name__"],"warnings":["something went awfuly wrong", "agaaaain"]}`,
			annotations.Annotations(map[string]error{
				"something went awfuly wrong": errors.New("something went awfuly wrong"),
				"agaaaain":                    errors.New("agaaaain"),
			}),
		},
		{
			`{"status":"success","data":["__name__"],"warnings":[]}`,
			annotations.Annotations(map[string]error{}),
		},
	}

	for _, tc := range testCases {
		mockRemote := MockRemote{
			mux: http.NewServeMux(),
		}

		mockRemote.mux.HandleFunc(remotestorage.DefaultLabelNamesPath, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(tc.response))
			panicOnError(err)
		})

		remoteSrv := httptest.NewServer(mockRemote.mux)

		sut := remotestorage.NewRemoteStorage(logg, config.RemoteConfig{Name: "test", Address: remoteSrv.URL}, func() time.Time { return frozenTime })
		_, annots, err := sut.LabelNames(context.Background())
		assert.NoError(t, err, "should have returned NO error")
		assert.Equal(
			t,
			tc.expected,
			annots,
			"annotations should match",
		)

		remoteSrv.Close()
	}
}
