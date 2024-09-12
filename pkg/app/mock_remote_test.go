package app_test

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/jademcosta/graviola/pkg/domain"
)

type MockRemote struct {
	mux                   *http.ServeMux
	seriesSetMap          map[string]mockRemoteRoute
	calledWith            []*http.Request
	fixedStatusCodeAnswer int
	responses             [][]byte
}

func NewMockRemote(seriesSetMap map[string]mockRemoteRoute) *MockRemote {
	mux := http.NewServeMux()

	mock := &MockRemote{
		mux:          mux,
		seriesSetMap: seriesSetMap,
	}

	mux.Handle("/", mock)

	return mock
}

func (mock *MockRemote) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if mock.fixedStatusCodeAnswer != 0 {
		w.WriteHeader(mock.fixedStatusCodeAnswer)
		return
	}

	mock.calledWith = append(mock.calledWith, r)
	err := r.ParseForm()
	panicOnError(err)

	remote, ok := mock.seriesSetMap[r.URL.Path]

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(remote.status)
	resp := remote.encodeResponse()
	_, err = w.Write(resp)
	panicOnError(err)
	mock.responses = append(mock.responses, resp)
}

type mockRemoteRoute struct {
	status     int
	series     *domain.GraviolaSeriesSet
	resultType string
}

func (mockRR *mockRemoteRoute) encodeResponse() []byte {
	builder := strings.Builder{}
	_, err := builder.WriteString(`{"status":"`)
	panicOnError(err)

	successfulResponse := mockRR.status >= 200 && mockRR.status < 300
	if successfulResponse {
		_, err = builder.WriteString("success")
	} else {
		_, err = builder.WriteString("error")
	}
	panicOnError(err)

	_, err = builder.WriteString(fmt.Sprintf(`","data":{"resultType":"%s","result":[`, mockRR.resultType))
	panicOnError(err)

	_, err = builder.WriteString(mockRR.metricsAsJSONString())
	panicOnError(err)

	_, err = builder.WriteString(`]}}`)
	panicOnError(err)

	return []byte(builder.String())
}

func (mockRR *mockRemoteRoute) metricsAsJSONString() string {
	builder := strings.Builder{}

	for _, serie := range mockRR.series.Series {
		_, err := builder.WriteString(`{"metric":`)
		panicOnError(err)

		lblsJSON, err := serie.Lbs.MarshalJSON()
		panicOnError(err)

		_, err = builder.Write(lblsJSON)
		panicOnError(err)

		_, err = builder.WriteString(`,"values":[`)
		panicOnError(err)

		for _, dtpt := range serie.Datapoints {
			_, err = builder.WriteString("[" + dtpt.Timestamp.String() + ",")
			panicOnError(err)

			_, err = builder.WriteString(`"` + dtpt.Value.String() + `"]`)
			panicOnError(err)
		}
		_, err = builder.WriteString("]},")
		panicOnError(err)
	}
	//,{"metric":{"__name__":"csecondseries","job":"prometheus","xlastlabel":"any value"},"value":[1702174837.986,"1"]}
	result := builder.String()
	return strings.TrimSuffix(result, ",")
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
