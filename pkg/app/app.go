package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	grafanaregexp "github.com/grafana/regexp"
	"github.com/jademcosta/graviola/pkg/api"
	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/graviolalog"
	"github.com/jademcosta/graviola/pkg/o11y"
	"github.com/jademcosta/graviola/pkg/queryengine"
	"github.com/jademcosta/graviola/pkg/remotestorage"
	"github.com/jademcosta/graviola/pkg/remotestoragegroup"
	"github.com/jademcosta/graviola/pkg/storageproxy"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/common/version"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/util/notifications"
	"github.com/prometheus/prometheus/web"
	api_v1 "github.com/prometheus/prometheus/web/api/v1"
)

type App struct {
	api       *api.GraviolaAPI
	logger    *slog.Logger
	metricz   *prometheus.Registry
	conf      config.GraviolaConfig // TODO: this is needed due to the api server configs
	cancelCtx context.CancelFunc
}

func NewApp(conf config.GraviolaConfig) *App {
	logger := graviolalog.NewLogger(conf.LogConf)
	metricRegistry := prometheus.NewRegistry()

	eng := queryengine.NewGraviolaQueryEngine(logger, metricRegistry, conf)

	storageGroups := initializeRemoteGroups(
		logger, metricRegistry, conf.StoragesConf.Groups, conf.QueryConf.TimeoutDuration())
	mainMergeStrategy := remotestoragegroup.MergeStrategyFactory(conf.StoragesConf.MergeConf.Strategy)
	graviolaStorage := storageproxy.NewGraviolaStorage(logger, storageGroups, mainMergeStrategy)

	apiV1 := createPrometheusAPI(eng, graviolaStorage, logger, metricRegistry)

	metricRegistry.MustRegister(
		collectors.NewBuildInfoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(
			collectors.WithGoCollectorRuntimeMetrics(collectors.GoRuntimeMetricsRule{Matcher: regexp.MustCompile("/.*")}),
		),
	)

	graviolaAPI := api.NewGraviolaAPI(conf.APIConf, logger, metricRegistry, apiV1)

	return &App{
		api:     graviolaAPI,
		logger:  logger,
		metricz: metricRegistry,
		conf:    conf,
	}
}

func (app *App) Start() {
	appCtx, cancelCtx := context.WithCancel(context.Background())
	app.cancelCtx = cancelCtx

	g := run.Group{}

	g.Add(func() error {
		return app.api.Start()
	}, func(_ error) {
		app.api.Stop()
		cancelCtx()
	})

	g.Add(func() error {
		signalsCh := make(chan os.Signal, 2)
		signal.Notify(signalsCh, syscall.SIGINT, syscall.SIGTERM)

		select {
		case s := <-signalsCh:
			app.logger.Info("received system signal", "signal", s.String())
		case <-appCtx.Done():
		}
		return nil
	}, func(_ error) {
		cancelCtx()
	})

	app.logger.Info("starting Graviola")
	err := g.Run()
	if err != nil {
		app.logger.Error("error after start", "error", err)
	}
}

func (app *App) Stop() {
	if app.cancelCtx != nil {
		app.logger.Info("stopping Graviola")
		app.cancelCtx()
	} else {
		app.logger.Error("no context cancel function registered")
	}
}

func alwaysReadyHandler(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f(w, r)
	}
}

func initializeRemoteGroups(
	logger *slog.Logger, metricz *prometheus.Registry, groupsConf []config.RemoteGroupsConfig,
	defaultQueryTimeout time.Duration,
) []storage.Querier {
	groups := make([]storage.Querier, 0, len(groupsConf))

	for _, groupConf := range groupsConf {
		failureStrategy := remotestoragegroup.QueryFailureStrategyFactory(groupConf.OnQueryFailStrategy)
		//FIXME: allow to configure this
		mergeStrategy := remotestoragegroup.MergeStrategyFactory(config.MergeStrategyAlwaysMerge)

		group := remotestoragegroup.NewRemoteGroup(
			logger,
			groupConf.Name,
			initializeRemotes(logger, metricz, groupConf.Servers, defaultQueryTimeout),
			failureStrategy,
			mergeStrategy,
		)
		groups = append(groups, o11y.NewQuerierO11y(metricz, groupConf.Name, "group", group))
	}

	return groups
}

func initializeRemotes(
	logger *slog.Logger, metricz *prometheus.Registry, remotesConf []config.RemoteConfig,
	defaultTimeout time.Duration,
) []storage.Querier {
	remotes := make([]storage.Querier, 0, len(remotesConf))

	for _, remoteConf := range remotesConf {
		remote := remotestorage.NewRemoteStorage(logger, remoteConf, time.Now, defaultTimeout)
		remotes = append(remotes, o11y.NewQuerierO11y(metricz, remoteConf.Name, "remote", remote))
	}

	return remotes
}

func createPrometheusAPI(
	queryEngine *queryengine.GraviolaQueryEngine,
	graviolaStorage *storageproxy.GraviolaStorage,
	logger *slog.Logger,
	metricRegistry *prometheus.Registry,
) *api_v1.API {

	//TODO: avoid all nils in the functions below. To avoid `panic`s
	return api_v1.NewAPI(
		queryEngine,
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
		false,          //remoteWriteEnabled
		nil,            // acceptRemoteWriteProtoMsgs []config.RemoteWriteProtoMsg,
		false,          //otlpEnabled
		false,          //otlpDeltaToCumulative
		false,          //otlpNativeDeltaIngestion
		false,          //ctZeroIngestionEnabled
	)
}
