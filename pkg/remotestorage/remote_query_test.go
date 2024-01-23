package remotestorage_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/graviolalog"
	"github.com/jademcosta/graviola/pkg/remotestorage"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/stretchr/testify/assert"
)

var logg *slog.Logger = graviolalog.NewLogger(config.LogConfig{Level: "error"})
var frozenTime = time.Now()

const defaultVectorAnswer = `{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"up","instance":"localhost:9090","job":"prometheus"},"value":[1702174837.986,"1"]}]}}`

type MockRemote struct {
	mux *http.ServeMux
}

func TestMarshalsTheQueryPayloadCorrectly(t *testing.T) {

	bodiesInstantQuery := make([]string, 0)
	bodiesRangeQuery := make([]string, 0)
	mockRemote := MockRemote{
		mux: http.NewServeMux(),
	}
	mockRemote.mux.HandleFunc(remotestorage.DefaultInstantQueryPath, func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err, "should not error reading body")
		bodiesInstantQuery = append(bodiesInstantQuery, string(body))
		bodiesRangeQuery = append(bodiesRangeQuery, "garbage")
		_, err = w.Write([]byte(defaultVectorAnswer))
		if err != nil {
			panic(err)
		}
	})

	mockRemote.mux.HandleFunc(remotestorage.DefaultRangeQueryPath, func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err, "should not error reading body")
		bodiesRangeQuery = append(bodiesRangeQuery, string(body))
		bodiesInstantQuery = append(bodiesInstantQuery, "garbage")
		_, err = w.Write([]byte(defaultVectorAnswer))
		if err != nil {
			panic(err)
		}
	})

	remoteSrv := httptest.NewServer(mockRemote.mux)
	defer remoteSrv.Close()

	testCases := []struct {
		matchers     []*labels.Matcher
		hints        *storage.SelectHints
		expected     string
		instantQuery bool
	}{
		{
			[]*labels.Matcher{
				labels.MustNewMatcher(labels.MatchEqual, "__name__", "value1"),
				labels.MustNewMatcher(labels.MatchRegexp, "label2", "value2.*"),
			},
			&storage.SelectHints{Start: 1234000, End: 5678000, Step: 7000},
			// Notice that the values are turned from milliseconds into seconds
			`end=5678&query={__name__="value1",label2=~"value2.*",}&start=1234&step=7`,
			false,
		},
		{
			[]*labels.Matcher{
				labels.MustNewMatcher(labels.MatchEqual, "labelName", "labelVal"),
			},
			&storage.SelectHints{},
			`query={labelName="labelVal",}`,
			true,
		},
		{
			[]*labels.Matcher{
				labels.MustNewMatcher(labels.MatchEqual, "labelName", "labelVal"),
			},
			&storage.SelectHints{Step: 33},
			`query={labelName="labelVal",}`,
			true,
		},
		{
			[]*labels.Matcher{
				labels.MustNewMatcher(labels.MatchEqual, "__name__", "value1"),
				labels.MustNewMatcher(labels.MatchRegexp, "label2", "value2.*"),
			},
			&storage.SelectHints{Start: 1234000, End: 5678000},
			`end=5678&query={__name__="value1",label2=~"value2.*",}&start=1234&step=30`,
			false,
		},
		{
			[]*labels.Matcher{
				labels.MustNewMatcher(labels.MatchEqual, "__name__", "value1"),
				labels.MustNewMatcher(labels.MatchRegexp, "label2", "value2.*"),
			},
			&storage.SelectHints{Start: 5678000, End: 5678000},
			`query={__name__="value1",label2=~"value2.*",}&time=5678`,
			true,
		},
	}

	sut := remotestorage.NewRemoteStorage(logg, config.RemoteConfig{Name: "test", Address: remoteSrv.URL}, func() time.Time { return frozenTime })

	for idx, tc := range testCases {
		result := sut.Select(context.Background(), true, tc.hints, tc.matchers...)
		assert.NotNil(t, result, "result should not be nil")
		time.Sleep(1 * time.Millisecond)
		var sentPayload string
		var err error
		if tc.instantQuery {
			sentPayload, err = url.PathUnescape(bodiesInstantQuery[idx])
		} else {
			sentPayload, err = url.PathUnescape(bodiesRangeQuery[idx])
		}
		assert.NoError(t, err, "should not error here")
		assert.Equal(t, tc.expected, sentPayload, "should have sent the correct payload")
	}
}

func TestUsesTheContextParameter(t *testing.T) {

	deadline := time.Now().Add(1 * time.Second)
	ctx, cancelFn := context.WithDeadline(context.Background(), deadline)
	defer cancelFn()

	mux := http.NewServeMux()
	mux.HandleFunc(remotestorage.DefaultInstantQueryPath, func(w http.ResponseWriter, r *http.Request) {

		// time.Sleep(200 * time.Millisecond)
		ddl, ok := ctx.Deadline()
		assert.True(t, ok, "should have a deadline set")
		assert.Equal(t, deadline, ddl, "should have the same deadline")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{},"value":[1702174837.986,"77"]}]}}`))
		panicOnError(err)
	})

	remoteSrv := httptest.NewServer(mux)
	defer remoteSrv.Close()

	matchers := []*labels.Matcher{
		labels.MustNewMatcher(labels.MatchEqual, "labelName", "labelVal"),
	}
	hints := &storage.SelectHints{}

	sut := remotestorage.NewRemoteStorage(logg, config.RemoteConfig{Name: "test", Address: remoteSrv.URL}, func() time.Time { return frozenTime })

	result := sut.Select(ctx, true, hints, matchers...)
	assert.NotNil(t, result, "result should not be nil")
	time.Sleep(50 * time.Millisecond)
}
