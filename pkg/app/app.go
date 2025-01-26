package app

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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	grafanaregexp "github.com/grafana/regexp"
	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/graviolalog"
	"github.com/jademcosta/graviola/pkg/http/httpmiddleware"
	"github.com/jademcosta/graviola/pkg/o11y"
	"github.com/jademcosta/graviola/pkg/queryengine"
	"github.com/jademcosta/graviola/pkg/remotestorage"
	"github.com/jademcosta/graviola/pkg/remotestoragegroup"
	"github.com/jademcosta/graviola/pkg/storageproxy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/route"
	"github.com/prometheus/common/version"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/util/notifications"
	"github.com/prometheus/prometheus/web"
	api_v1 "github.com/prometheus/prometheus/web/api/v1"
)

const remoteWriteEnabled = false
const otlpEnabled = false

type App struct {
	api     *api_v1.API
	router  *chi.Mux
	logger  *slog.Logger
	metricz *prometheus.Registry
	Srv     *http.Server
	conf    config.GraviolaConfig // TODO: this is needed due to the api server configs
}

func NewApp(conf config.GraviolaConfig) *App {
	logger := graviolalog.NewLogger(conf.LogConf)
	metricRegistry := prometheus.NewRegistry()

	eng := queryengine.NewGraviolaQueryEngine(logger, metricRegistry, conf)

	storageGroups := initializeGroups(logger, metricRegistry, conf.StoragesConf.Groups)
	mainMergeStrategy := remotestoragegroup.MergeStrategyFactory(conf.StoragesConf.MergeConf.Strategy)
	graviolaStorage := storageproxy.NewGraviolaStorage(logger, storageGroups, mainMergeStrategy)

	//TODO: avoid all nils in the functions below. This is to avoid `panic`s
	apiV1 := api_v1.NewAPI(
		eng,
		graviolaStorage,
		nil, // storage.Appendable // seems to be Ok to be nil
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
		logger,
		nil,                             // func(context.Context) RulesRetriever
		100,                             //TODO: allow config (remoteReadSampleLimit)
		10,                              //TODO: allow config (remoteReadConcurrencyLimit)
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

		func() []notifications.Notification {
			return nil
		}, //notificationsGetter, to get notifications to show on the UI
		func() (<-chan notifications.Notification, func(), bool) {
			noNotificationStreamAnswer := false
			return nil, func() {}, noNotificationStreamAnswer
		}, // notificationsSub, to get SSE notifications, live

		metricRegistry, // gatherer prometheus.Gatherer
		metricRegistry, // registerer prometheus.Registerer
		nil,            // statsRenderer StatsRenderer
		remoteWriteEnabled,
		nil, // acceptRemoteWriteProtoMsgs []config.RemoteWriteProtoMsg,
		otlpEnabled,
	)

	metricRegistry.MustRegister(
		collectors.NewBuildInfoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(
			collectors.WithGoCollectorRuntimeMetrics(collectors.GoRuntimeMetricsRule{Matcher: regexp.MustCompile("/.*")}),
		),
	)

	router := chi.NewRouter()

	router.Use(httpmiddleware.NewLoggingMiddleware(logger))
	router.Use(httpmiddleware.NewMetricsMiddleware(metricRegistry))
	router.Use(middleware.Recoverer)

	router.Get("/metrics", promhttp.HandlerFor(metricRegistry, promhttp.HandlerOpts{Registry: metricRegistry}).ServeHTTP)
	router.Get("/healthy", alwaysSuccessfulHandler)
	router.Get("/ready", alwaysSuccessfulHandler)

	subRouter := route.New()
	subRouter = subRouter.WithPrefix("/api/v1")
	apiV1.Register(subRouter)
	router.Handle("/*", subRouter)

	srv := &http.Server{Addr: fmt.Sprintf(":%d", conf.APIConf.Port), Handler: router} //TODO: extract and allow config

	return &App{
		api:     apiV1,
		router:  router,
		logger:  logger,
		metricz: metricRegistry,
		Srv:     srv,
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
	app.logger.Error("listenandserve exited", "error", fmt.Errorf("on serving HTTP: %w", app.Srv.ListenAndServe()))
}

func (app *App) Stop() {
	//TODO: should this timeout be a config on its own?
	ctx, cancelFn := context.WithTimeout(context.Background(), app.conf.QueryConf.TimeoutDuration())
	defer cancelFn()

	app.logger.Info("shutting down")

	err := app.Srv.Shutdown(ctx)
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

func initializeGroups(
	logger *slog.Logger, metricz *prometheus.Registry, groupsConf []config.GroupsConfig,
) []storage.Querier {
	groups := make([]storage.Querier, 0, len(groupsConf))

	for _, groupConf := range groupsConf {
		failureStrategy := remotestoragegroup.QueryFailureStrategyFactory(groupConf.OnQueryFailStrategy)
		//TODO: allow to configure this
		mergeStrategy := remotestoragegroup.MergeStrategyFactory(config.MergeStrategyAlwaysMerge)

		group := remotestoragegroup.NewGroup(
			logger, groupConf.Name,
			initializeRemotes(logger, metricz, groupConf.Servers), failureStrategy, mergeStrategy)
		groups = append(groups, o11y.NewQuerierO11y(metricz, groupConf.Name, "group", group))
	}

	return groups
}

func initializeRemotes(
	logger *slog.Logger, metricz *prometheus.Registry, remotesConf []config.RemoteConfig,
) []storage.Querier {
	remotes := make([]storage.Querier, 0, len(remotesConf))

	for _, remoteConf := range remotesConf {
		remote := remotestorage.NewRemoteStorage(logger, remoteConf, time.Now)
		remotes = append(remotes, o11y.NewQuerierO11y(metricz, remoteConf.Name, "remote", remote))
	}

	return remotes
}

func alwaysSuccessfulHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
