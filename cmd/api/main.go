package main

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	grafanaregexp "github.com/grafana/regexp"
	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/fakes"
	"github.com/jademcosta/graviola/pkg/graviolalog"
	"github.com/jademcosta/graviola/pkg/querytracker"
	"github.com/jademcosta/graviola/pkg/storageproxy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/route"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/util/stats"
	api_v1 "github.com/prometheus/prometheus/web/api/v1"
)

func main() {

	logger := graviolalog.NewLogger(config.LogConfig{Level: "info"})

	graviolaStorage := &storageproxy.GraviolaStorage{Logg: logger}
	graviolaExemplarQueryable := &storageproxy.GraviolaExemplarQueryable{}

	eng := promql.NewEngine(promql.EngineOpts{
		Timeout:            1 * time.Minute,
		MaxSamples:         10000,
		LookbackDelta:      5 * time.Minute,
		EnableAtModifier:   true,
		ActiveQueryTracker: querytracker.NewGraviolaQueryTracker(2),
	})

	dbDir := ""
	enableAdmin := false
	remoteReadLimits := 0 //TODO: add this to todo list, it has to do with remote read capabilities
	isAgent := false
	corsOrigin, err := grafanaregexp.Compile(".*")
	if err != nil {
		panic("error on CORS regex")
	}
	buildInfo := &api_v1.PrometheusVersion{}
	metricRegistry := prometheus.NewRegistry()

	remoteWriteEnabled := false
	otlpEnabled := false

	apiV1 := api_v1.NewAPI(
		eng,
		graviolaStorage,
		nil, //Appender, seems to be Ok to be nil
		graviolaExemplarQueryable,
		nil,
		nil,
		nil,
		nil, //TODO: return a valid config
		make(map[string]string, 0),
		api_v1.GlobalURLOptions{},
		testReady,
		&fakes.FakeTSDBAdminStats{}, //Can this be nil?
		dbDir,
		enableAdmin,
		graviolalog.AdaptToGoKitLogger(logger),
		rulesRetrvFn,
		remoteReadLimits,
		remoteReadLimits,
		remoteReadLimits,
		isAgent,
		corsOrigin,
		runtimeInfo,
		buildInfo,
		metricRegistry, //FIXME: is it a gatherer?
		metricRegistry,
		statsRender,
		remoteWriteEnabled,
		otlpEnabled,
	)

	router := route.New()

	metricRegistry.MustRegister(
		collectors.NewBuildInfoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(
			collectors.WithGoCollectorRuntimeMetrics(collectors.GoRuntimeMetricsRule{Matcher: regexp.MustCompile("/.*")}),
		),
	)
	router.Get("/metrics", promhttp.HandlerFor(metricRegistry, promhttp.HandlerOpts{Registry: metricRegistry}).ServeHTTP)

	router = router.WithPrefix("/api/v1")

	apiV1.Register(router)

	var srv *http.Server = &http.Server{Addr: fmt.Sprintf(":%d", 8081), Handler: router}
	logger.Info("Starting server...")
	fmt.Println(fmt.Errorf("on serving HTTP: %w", srv.ListenAndServe()))
}

func testReady(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f(w, r)
	}
}

func rulesRetrvFn(context.Context) api_v1.RulesRetriever {
	// panic("should not be called")
	return nil //TODO: create a fake that `panics`
}

func runtimeInfo() (api_v1.RuntimeInfo, error) {
	// panic("should not be called")
	return api_v1.RuntimeInfo{}, nil //TODO: this is a first thing to do after making it work
}

func statsRender(context.Context, *stats.Statistics, string) stats.QueryStats {
	// panic("should not be called")
	return nil
}
