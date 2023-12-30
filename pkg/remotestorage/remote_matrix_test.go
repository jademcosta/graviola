package remotestorage_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/jademcosta/graviola/pkg/remotestorage"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/util/annotations"
	"github.com/stretchr/testify/assert"
)

func TestParsesMatrixResponseCorrectlyWithOrderedLabelsAndSeries(t *testing.T) {

	defaultHints := storage.SelectHints{}
	defaultMatchers := []*labels.Matcher{{Type: labels.MatchEqual, Name: "abc", Value: "123"}}

	var response *string

	mockRemote := MockRemote{
		mux: http.NewServeMux(),
	}
	mockRemote.mux.HandleFunc(remotestorage.DefaultInstantQueryPath, func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(*response))
		if err != nil {
			panic(err)
		}
	})

	remoteSrv := httptest.NewServer(mockRemote.mux)
	defer remoteSrv.Close()

	cases := loadMatrixCases()

	testCases := []struct {
		answer      string
		seriesCount int
		annots      annotations.Annotations
	}{
		{
			`{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"__name__":"up","instance":"localhost:9090","job":"prometheus"},"values":[[1701635580,"1.2"], [1701635610,"1.6"]]}]}}`,
			1,
			nil,
		},
		{
			`{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"__name__":"up","instance":"localhost:9090","job":"prometheus"},"values":[[1701635580,"1"],[1701635610,"1"],[1701635640,"1"],[1701635670,"1"],[1701635700,"1"],[1701635730,"1"],[1701635760,"1"],[1701635790,"1"],[1701635820,"1"],[1701635850,"1"],[1701635880,"1"],[1701635910,"1"],[1701635940,"1"],[1701635970,"1"],[1701636000,"1"],[1701636030,"1"],[1701636060,"1"],[1701636090,"1"],[1701636120,"1"],[1701636150,"1"],[1701636180,"1"],[1701636210,"1"],[1701636240,"1"],[1701636270,"1"],[1701636300,"1"],[1701636330,"1"],[1701636360,"1"],[1701636390,"1"],[1701636420,"1"],[1701636450,"1"],[1701636480,"1"],[1701636510,"1"],[1701636540,"1"],[1701636570,"1"],[1701636600,"1"],[1701636630,"1"],[1701636660,"1"],[1701636690,"1"],[1701636720,"1"],[1701636750,"1"],[1701636780,"1"],[1701636810,"1"],[1701636840,"1"],[1701636870,"1"],[1701636900,"1"],[1701636930,"1"],[1701636960,"1"],[1701636990,"1"],[1701637020,"1"],[1701637050,"1"],[1701637080,"1"],[1701637110,"1"],[1701637140,"1"],[1701637170,"1"],[1701637200,"1"],[1701637230,"1"],[1701637260,"1"],[1701637290,"1"],[1701637320,"1"],[1701637350,"1"],[1701637380,"1"],[1701637410,"1"],[1701637440,"1"],[1701637470,"1"],[1701637500,"1"],[1701637530,"1"],[1701637560,"1"],[1701637590,"1"],[1701637620,"1"],[1701637650,"1"],[1701637680,"1"],[1701637710,"1"],[1701637740,"1"],[1701637770,"1"],[1701637800,"1"],[1701637830,"1"],[1701637860,"1"],[1701637890,"1"],[1701637920,"1"],[1701637950,"1"],[1701637980,"1"],[1701638010,"1"],[1701638040,"1"],[1701638070,"1"],[1701638100,"1"],[1701638130,"1"],[1701638160,"1"],[1701638190,"1"],[1701638220,"1"],[1701638250,"1"],[1701638280,"1"],[1701638310,"1"],[1701638340,"1"],[1701638370,"1"],[1701638400,"1"],[1701638430,"1"],[1701638460,"1"],[1701638490,"1"],[1701638520,"1"],[1701638550,"1"],[1701638580,"1"],[1701638610,"1"],[1701638640,"1"],[1701638670,"1"],[1701638700,"1"],[1701638730,"1"],[1701638760,"1"],[1701638790,"1"],[1701638820,"1"],[1701638850,"1"],[1701638880,"1"],[1701638910,"1"],[1701638940,"1"],[1701638970,"1"],[1701639000,"1"]]},{"metric":{"__name__":"up","instance":"localhost:9100","job":"node"},"values":[[1701635580,"1"],[1701635610,"1"],[1701635640,"1"],[1701635670,"1"],[1701635700,"1"],[1701635730,"1"],[1701635760,"1"],[1701635790,"1"],[1701635820,"1"],[1701635850,"1"],[1701635880,"1"],[1701635910,"1"],[1701635940,"1"],[1701635970,"1"],[1701636000,"1"],[1701636030,"1"],[1701636060,"1"],[1701636090,"1"],[1701636120,"1"],[1701636150,"1"],[1701636180,"1"],[1701636210,"1"],[1701636240,"1"],[1701636270,"1"],[1701636300,"1"],[1701636330,"1"],[1701636360,"1"],[1701636390,"1"],[1701636420,"1"],[1701636450,"1"],[1701636480,"1"],[1701636510,"1"],[1701636540,"1"],[1701636570,"1"],[1701636600,"1"],[1701636630,"1"],[1701636660,"1"],[1701636690,"1"],[1701636720,"1"],[1701636750,"1"],[1701636780,"1"],[1701636810,"1"],[1701636840,"1"],[1701636870,"1"],[1701636900,"1"],[1701636930,"1"],[1701636960,"1"],[1701636990,"1"],[1701637020,"1"],[1701637050,"1"],[1701637080,"1"],[1701637110,"1"],[1701637140,"1"],[1701637170,"1"],[1701637200,"1"],[1701637230,"1"],[1701637260,"1"],[1701637290,"1"],[1701637320,"1"],[1701637350,"1"],[1701637380,"1"],[1701637410,"1"],[1701637440,"1"],[1701637470,"1"],[1701637500,"1"],[1701637530,"1"],[1701637560,"1"],[1701637590,"1"],[1701637620,"1"],[1701637650,"1"],[1701637680,"1"],[1701637710,"1"],[1701637740,"1"],[1701637770,"1"],[1701637800,"1"],[1701637830,"1"],[1701637860,"1"],[1701637890,"1"],[1701637920,"1"],[1701637950,"1"],[1701637980,"1"],[1701638010,"1"],[1701638040,"1"],[1701638070,"1"],[1701638100,"1"],[1701638130,"1"],[1701638160,"1"],[1701638190,"1"],[1701638220,"1"],[1701638250,"1"],[1701638280,"1"],[1701638310,"1"],[1701638340,"1"],[1701638370,"1"],[1701638400,"1"],[1701638430,"1"],[1701638460,"1"],[1701638490,"1"],[1701638520,"1"],[1701638550,"1"],[1701638580,"1"],[1701638610,"1"],[1701638640,"1"],[1701638670,"1"],[1701638700,"1"],[1701638730,"1"],[1701638760,"1"],[1701638790,"1"],[1701638820,"1"],[1701638850,"1"],[1701638880,"1"],[1701638910,"1"],[1701638940,"1"],[1701638970,"1"],[1701639000,"14"]]}]}}`,
			2,
			nil,
		},
		{
			cases[1], //TODO: use case 0 too?
			1,
			annotations.New().Add(errors.New("warnings: [PromQL info: input to histogram_quantile needed to be fixed for monotonicity (and may give inaccurate results) for metric name \"\" (1:25)]")),
		},
		{
			cases[2],
			2,
			nil,
		},
	}

	sut := remotestorage.NewRemoteStorage(logg, config.RemoteConfig{Name: "test", Address: remoteSrv.URL}, func() time.Time { return frozenTime })

	for _, tc := range testCases {

		series := extractSeriesFrom(t, tc.answer)

		expectedResult := &domain.GraviolaSeriesSet{
			Series: series,
		}

		if len(tc.annots) > 0 {
			expectedResult.Annots = tc.annots
		}

		response = &tc.answer
		result := sut.Select(true, &defaultHints, defaultMatchers...)
		assert.Lenf(t, result.(*domain.GraviolaSeriesSet).Series, tc.seriesCount, "should have %d series", tc.seriesCount)

		for idx, resultSeries := range series {
			assert.Equal(t, series[idx], resultSeries, "series should match")
		}

		assert.Equal(t, expectedResult, result, "result should be correct")
	}
}

// Reading reflect.DeepEqual docs, NaN values are NEVER equal to another NaN. As such,
// they cannot be compared.
func TestParsesMatrixResponseCorrectlyWithNaNs(t *testing.T) {

	defaultHints := storage.SelectHints{}
	defaultMatchers := []*labels.Matcher{{Type: labels.MatchEqual, Name: "abc", Value: "123"}}

	response := `{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"handler":"/"},"values":[[1701635640,"NaN"]]}]}}`

	mockRemote := MockRemote{
		mux: http.NewServeMux(),
	}
	mockRemote.mux.HandleFunc(remotestorage.DefaultInstantQueryPath, func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(response))
		if err != nil {
			panic(err)
		}
	})

	remoteSrv := httptest.NewServer(mockRemote.mux)
	defer remoteSrv.Close()

	sut := remotestorage.NewRemoteStorage(logg, config.RemoteConfig{Name: "test", Address: remoteSrv.URL}, func() time.Time { return frozenTime })

	result := sut.Select(true, &defaultHints, defaultMatchers...)
	assert.Lenf(t, result.(*domain.GraviolaSeriesSet).Series, 1, "should have 1 serie")

	resultParsed := result.(*domain.GraviolaSeriesSet)

	assert.Equal(t, "handler", resultParsed.Series[0].Lbs[0].Name, "label name should have been parsed")
	assert.Equal(t, "/", resultParsed.Series[0].Lbs[0].Value, "label value should have been parsed")

	assert.Equal(t, model.Time(1701635640000), resultParsed.Series[0].Datapoints[0].Timestamp, "timestamp should have been parsed")
	assert.True(t, math.IsNaN(float64(resultParsed.Series[0].Datapoints[0].Value)), "NaN value should have been parsed")
}

func loadMatrixCases() []string {

	answer := make([]string, 0)

	content, err := os.ReadFile("../../test/matrix_answer_with_nan.json")
	if err != nil {
		panic(fmt.Errorf("error reading file: %w", err))
	}
	answer = append(answer, string(content))

	content, err = os.ReadFile("../../test/matrix_answer_with_warning_without_label.json")
	if err != nil {
		panic(fmt.Errorf("error reading file: %w", err))
	}
	answer = append(answer, string(content))

	content, err = os.ReadFile("../../test/matrix_answer.json")
	if err != nil {
		panic(fmt.Errorf("error reading file: %w", err))
	}
	answer = append(answer, string(content))

	return answer
}

func extractSeriesFrom(t *testing.T, remoteAnswer string) []*domain.GraviolaSeries {

	var tempMap map[string]*json.RawMessage
	err := json.Unmarshal([]byte(remoteAnswer), &tempMap)
	assert.NoError(t, err, "should not return error")

	err = json.Unmarshal(*tempMap["data"], &tempMap)
	assert.NoError(t, err, "should not return error")

	var metrics model.Matrix
	err = json.Unmarshal(*tempMap["result"], &metrics)
	assert.NoError(t, err, "should not return error")

	series := make([]*domain.GraviolaSeries, 0)

	for _, metric := range metrics {
		datapoints := make([]model.SamplePair, 0)
		lbls := make(map[string]string, 0)

		for _, sample := range metric.Values {

			datapoints = append(datapoints, model.SamplePair{
				Timestamp: sample.Timestamp,
				Value:     sample.Value,
			})
		}

		for lblName, lblValue := range metric.Metric {
			lbls[string(lblName)] = string(lblValue)
		}

		series = append(series, &domain.GraviolaSeries{
			Lbs:        labels.FromMap(lbls),
			Datapoints: datapoints,
		})
	}
	return series
}
