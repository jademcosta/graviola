package application

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	grafanaregexp "github.com/grafana/regexp"
	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/graviolalog"
	"github.com/jademcosta/graviola/pkg/querytracker"
	"github.com/jademcosta/graviola/pkg/remotestorage"
	"github.com/jademcosta/graviola/pkg/remotestoragegroup"
	"github.com/jademcosta/graviola/pkg/storageproxy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/route"
	"github.com/prometheus/common/version"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/web"
	api_v1 "github.com/prometheus/prometheus/web/api/v1"
)

const remoteWriteEnabled = false
const otlpEnabled = false

type App struct {
	api     *api_v1.API
	logger  *slog.Logger
	metricz *prometheus.Registry
	srv     *http.Server
	conf    config.GraviolaConfig // TODO: this is needed due to the api server configs
}

func NewApp(conf config.GraviolaConfig) *App {
	logger := graviolalog.NewLogger(conf.LogConf)
	metricRegistry := prometheus.NewRegistry()

	eng := promql.NewEngine(promql.EngineOpts{
		Timeout:            conf.ApiConf.TimeoutDuration(),
		MaxSamples:         10000, //TODO: add config for all these
		LookbackDelta:      5 * time.Minute,
		EnableAtModifier:   true,
		ActiveQueryTracker: querytracker.NewGraviolaQueryTracker(2),
	})

	graviolaStorage := storageproxy.NewGraviolaStorage(logger, initializeGroups(logger, conf.StoragesConf.Groups))

	apiV1 := api_v1.NewAPI(
		eng,
		graviolaStorage,
		nil, //TODO: Appender, seems to be Ok to be nil
		&storageproxy.GraviolaExemplarQueryable{},
		nil,                        // func(context.Context) ScrapePoolsRetriever
		nil,                        // func(context.Context) TargetRetriever
		nil,                        // func(context.Context) AlertmanagerRetriever
		nil,                        // func() config.Config //TODO: this might panic if config endpoint is hit
		make(map[string]string, 0), // This is used on the flags endpoint //TODO: add the config file flag
		api_v1.GlobalURLOptions{},  // This is used on the targets endpoint
		alwaysReadyHandler,         // TODO: do I need to use this one? It is used to prevent calling certain endpoints without being ready
		nil,                        // TSDBAdminStats
		"",                         // dbDir string
		false,                      // enableAdmin bool
		graviolalog.AdaptToGoKitLogger(logger),
		nil,                             // func(context.Context) RulesRetriever
		1,                               //TODO: allow config (remoteReadSampleLimit)
		1,                               //TODO: allow config (remoteReadConcurrencyLimit)
		1024,                            //TODO: allow config (remoteReadMaxBytesInFrame) (currently 1KB)
		false,                           // isAgent bool - If this is set to true the query endpoints will not work
		grafanaregexp.MustCompile(".*"), // corsOrigin *regexp.Regexp,
		nil,                             // runtimeInfo func() (RuntimeInfo, error)
		&web.PrometheusVersion{
			Version:   version.Version,
			Revision:  version.Revision,
			Branch:    version.Branch,
			BuildUser: version.BuildUser,
			BuildDate: version.BuildDate,
			GoVersion: version.GoVersion,
		}, // buildInfo *PrometheusVersion
		metricRegistry, // gatherer prometheus.Gatherer
		metricRegistry, // registerer prometheus.Registerer
		nil,            // statsRenderer StatsRenderer
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

	srv := &http.Server{Addr: fmt.Sprintf(":%d", conf.ApiConf.Port), Handler: router} //TODO: extract and allow config

	return &App{
		api:     apiV1,
		logger:  logger,
		metricz: metricRegistry,
		srv:     srv,
		conf:    conf,
	}
}

func (app *App) Start() {
	go func() {
		signalsCh := make(chan os.Signal, 2)
		signal.Notify(signalsCh, syscall.SIGINT, syscall.SIGTERM)

		s := <-signalsCh
		app.logger.Info("received signal", "signal", s.String())
		app.Stop()
	}()

	app.logger.Info("Starting server...")
	fmt.Println(fmt.Errorf("on serving HTTP: %w", app.srv.ListenAndServe()))
}

func (app *App) Stop() {
	ctx, cancelFn := context.WithTimeout(context.Background(), app.conf.ApiConf.TimeoutDuration())
	defer cancelFn()

	app.logger.Info("shutting down")

	err := app.srv.Shutdown(ctx)
	if err != nil {
		app.logger.Error("error when sutting down", "error", err)
	}
	app.logger.Info("shutdown finished")
}

func alwaysReadyHandler(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f(w, r)
	}
}

func initializeGroups(logger *slog.Logger, groupsConf []config.GroupsConfig) []storage.Querier {
	groups := make([]storage.Querier, 0, len(groupsConf))

	for _, groupConf := range groupsConf {
		groups = append(groups, remotestoragegroup.NewGroup(logger, groupConf.Name, initializeRemotes(logger, groupConf.Servers)))
	}

	return groups
}

func initializeRemotes(logger *slog.Logger, remotesConf []config.RemoteConfig) []storage.Querier {
	remotes := make([]storage.Querier, 0, len(remotesConf))

	for _, remoteConf := range remotesConf {
		remotes = append(remotes, remotestorage.NewRemoteStorage(logger, remoteConf, time.Now))
	}

	return remotes
}
