package remotestorage_test

import (
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/jademcosta/graviola/pkg/remotestorage"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/stretchr/testify/assert"
)

func TestParsesVectorResponseCorrectlyWithOrderedLabelsAndSeries(t *testing.T) {

	defaultHints := storage.SelectHints{}
	defaultMatchers := []*labels.Matcher{{Type: labels.MatchEqual, Name: "abc", Value: "123"}}

	var response *string

	mockRemote := MockRemote{
		mux: http.NewServeMux(),
	}
	mockRemote.mux.HandleFunc(remotestorage.DefaultInstantQueryPath, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(*response))
	})

	remoteSrv := httptest.NewServer(mockRemote.mux)
	defer remoteSrv.Close()

	testCases := []struct {
		answer         string
		expectedResult storage.SeriesSet
		seriesCount    int
	}{
		{
			defaultVectorAnswer,
			&domain.GraviolaSeriesSet{
				Series: []*domain.GraviolaSeries{
					{
						Lbs:  labels.FromStrings("__name__", "up", "instance", "localhost:9090", "job", "prometheus"),
						Ts:   []int64{1702174837986},
						Vals: []float64{1.0},
					},
				},
			},
			1,
		},
		{
			`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"xlastseries","aalabel1":"somelocal","job":"prometheus"},"value":[1702174837.986,"1"]},{"metric":{"__name__":"afirstseries","instance":"localhost:9090","job":"prometheus"},"value":[1702174837.986,"1"]},{"metric":{"__name__":"csecondseries","job":"prometheus","xlastlabel":"any value"},"value":[1702174837.986,"1"]}]}}`,
			&domain.GraviolaSeriesSet{
				Series: []*domain.GraviolaSeries{
					{
						Lbs:  labels.FromStrings("__name__", "afirstseries", "instance", "localhost:9090", "job", "prometheus"),
						Ts:   []int64{1702174837986},
						Vals: []float64{1.0},
					},
					{
						Lbs:  labels.FromStrings("__name__", "csecondseries", "job", "prometheus", "xlastlabel", "any value"),
						Ts:   []int64{1702174837986},
						Vals: []float64{1.0},
					},
					{
						Lbs:  labels.FromStrings("__name__", "xlastseries", "aalabel1", "somelocal", "job", "prometheus"),
						Ts:   []int64{1702174837986},
						Vals: []float64{1.0},
					},
				},
			},
			3,
		},
		{
			`{"status":"success","data":{"resultType":"vector","result":[{"metric":{},"value":[1702174837.986,"77"]}]}}`,
			&domain.GraviolaSeriesSet{
				Series: []*domain.GraviolaSeries{
					{
						Lbs:  labels.FromStrings(),
						Ts:   []int64{1702174837986},
						Vals: []float64{77.0},
					},
				},
			},
			1,
		},
		{
			`{"status":"success","data":{"resultType":"vector","result":[{"metric":{},"value":[1702174837.986,"77"]},{"metric":{"aaa":"111"},"value":[1702174837,"81.1"]}]}}`,
			&domain.GraviolaSeriesSet{
				Series: []*domain.GraviolaSeries{
					{
						Lbs:  labels.FromStrings(),
						Ts:   []int64{1702174837986},
						Vals: []float64{77.0},
					},
					{
						Lbs:  labels.FromStrings("aaa", "111"),
						Ts:   []int64{1702174837000},
						Vals: []float64{81.1},
					},
				},
			},
			2,
		},
		{
			`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"xlastlabel":"any value"},"value":[1.111,"1"]},{"metric":{"job":"prometheus"},"value":[1.111,"1"]},{"metric":{"job":"aaa"},"value":[1.111,"1"]}]}}`,
			&domain.GraviolaSeriesSet{
				Series: []*domain.GraviolaSeries{
					{
						Lbs:  labels.FromStrings("job", "aaa"),
						Ts:   []int64{1111},
						Vals: []float64{1.0},
					},
					{
						Lbs:  labels.FromStrings("job", "prometheus"),
						Ts:   []int64{1111},
						Vals: []float64{1.0},
					},
					{
						Lbs:  labels.FromStrings("xlastlabel", "any value"),
						Ts:   []int64{1111},
						Vals: []float64{1.0},
					},
				},
			},
			3,
		},
		{
			`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"xxmiddle"},"value":[1.111,"1"]},{"metric":{"__name__":"xxxlast","afirstlabelexceptnot":"prometheus"},"value":[1.111,"1"]},{"metric":{"__name__":"xfirst","xlastlabel":"any value"},"value":[1.111,"1"]}]}}`,
			&domain.GraviolaSeriesSet{
				Series: []*domain.GraviolaSeries{ //__name__ is the first label because labels are sorted
					{
						Lbs:  labels.FromStrings("__name__", "xfirst", "xlastlabel", "any value"),
						Ts:   []int64{1111},
						Vals: []float64{1.0},
					},
					{
						Lbs:  labels.FromStrings("__name__", "xxmiddle"), //The label comparison is made on the smallest number of labels
						Ts:   []int64{1111},
						Vals: []float64{1.0},
					},
					{
						Lbs:  labels.FromStrings("__name__", "xxxlast", "afirstlabelexceptnot", "prometheus"),
						Ts:   []int64{1111},
						Vals: []float64{1.0},
					},
				},
			},
			3,
		},
	}

	sut := remotestorage.NewRemoteStorage(logg, config.RemoteConfig{Name: "test", Address: remoteSrv.URL}, func() time.Time { return frozenTime })

	for _, tc := range testCases {
		response = &tc.answer
		result := sut.Select(true, &defaultHints, defaultMatchers...)
		assert.Lenf(t, result.(*domain.GraviolaSeriesSet).Series, tc.seriesCount, "should have %d series", tc.seriesCount)
		assert.Equal(t, tc.expectedResult, result, "result should be correct")
	}
}

func TestParsesVectorResponseCorrectlyWithNaN(t *testing.T) {

	defaultHints := storage.SelectHints{}
	defaultMatchers := []*labels.Matcher{{Type: labels.MatchEqual, Name: "abc", Value: "123"}}

	response := `{"status":"success","data":{"resultType":"vector","result":[{"metric":{},"value":[17024837.986,"NaN"]}]}}`

	mockRemote := MockRemote{
		mux: http.NewServeMux(),
	}
	mockRemote.mux.HandleFunc(remotestorage.DefaultInstantQueryPath, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(response))
	})

	remoteSrv := httptest.NewServer(mockRemote.mux)
	defer remoteSrv.Close()

	sut := remotestorage.NewRemoteStorage(logg, config.RemoteConfig{Name: "test", Address: remoteSrv.URL}, func() time.Time { return frozenTime })

	result := sut.Select(true, &defaultHints, defaultMatchers...)
	resultParsed := result.(*domain.GraviolaSeriesSet)

	assert.Lenf(t, resultParsed.Series, 1, "should have 1 serie")

	assert.Len(t, resultParsed.Series[0].Lbs, 0, "should have no labels")

	assert.Equal(t, int64(17024837986), resultParsed.Series[0].Ts[0], "timestamp should have been parsed")
	assert.True(t, math.IsNaN(resultParsed.Series[0].Vals[0]), "NaN value should have been parsed")
}
