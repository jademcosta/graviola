package application

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/grafana/regexp"
	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/graviolalog"
	"github.com/jademcosta/graviola/pkg/querytracker"
	"github.com/jademcosta/graviola/pkg/remotestorage"
	"github.com/jademcosta/graviola/pkg/remotestoragegroup"
	"github.com/jademcosta/graviola/pkg/storageproxy"
	"github.com/prometheus/client_golang/prometheus"
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
		nil,                      // func(context.Context) RulesRetriever
		1,                        //TODO: allow config (remoteReadSampleLimit)
		1,                        //TODO: allow config (remoteReadConcurrencyLimit)
		1024,                     //TODO: allow config (remoteReadMaxBytesInFrame) (currently 1KB)
		false,                    // isAgent bool - If this is set to true the query endpoints will not work
		regexp.MustCompile(".*"), // corsOrigin *regexp.Regexp,
		nil,                      // runtimeInfo func() (RuntimeInfo, error)
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

	return &App{
		api:     apiV1,
		logger:  logger,
		metricz: metricRegistry,
	}
}

func (app *App) Start() {

}

func (app *App) Stop() {

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
