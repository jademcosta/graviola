package queryengine_test

import (
	"os"
	"testing"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/graviolalog"
	"github.com/jademcosta/graviola/pkg/queryengine"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/promql"
)

func TestGraviolaQueryEngineComplyWithPromqlQueryEngine(_ *testing.T) {
	logger := graviolalog.NewNoopLogger()
	registry := prometheus.NewRegistry()

	confContent, err := os.ReadFile("../../config_example.yml")
	if err != nil {
		panic(err)
	}
	conf := config.MustParse(confContent)

	dummyFunc := func(_ promql.QueryEngine) {}

	sut := queryengine.NewGraviolaQueryEngine(logger, registry, conf)

	dummyFunc(sut)
}
