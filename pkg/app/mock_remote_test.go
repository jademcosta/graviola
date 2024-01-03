package app_test

import (
	"net/http"
)

type MockRemote struct {
	mux                   *http.ServeMux
	handlerMap            map[string]http.HandlerFunc
	calledWith            []*http.Request
	fixedStatusCodeAnswer int
}

func NewMockRemote(handlerMap map[string]http.HandlerFunc) *MockRemote {
	mux := http.NewServeMux()

	mock := &MockRemote{
		mux:        mux,
		handlerMap: handlerMap,
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

	h, ok := mock.handlerMap[r.URL.Path]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	h.ServeHTTP(w, r)
}
