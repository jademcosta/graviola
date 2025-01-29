package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/http/httpmiddleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/route"
)

type registerer interface {
	Register(*route.Router)
}

type GraviolaAPI struct {
	conf                config.APIConfig
	logger              *slog.Logger
	metricRegistry      *prometheus.Registry
	prometheusNativeAPI registerer
	srv                 *http.Server
	router              *chi.Mux
}

func NewGraviolaAPI(
	conf config.APIConfig,
	logger *slog.Logger,
	metricRegistry *prometheus.Registry,
	prometheusNativeAPI registerer,
) *GraviolaAPI {
	api := &GraviolaAPI{
		conf:                conf,
		logger:              logger.With("component", "api"),
		metricRegistry:      metricRegistry,
		prometheusNativeAPI: prometheusNativeAPI,
	}

	api.createRoutes()

	return api
}

func (api *GraviolaAPI) Start() error {
	return api.srv.ListenAndServe()
}

func (api *GraviolaAPI) Stop() {
	//TODO: should this timeout be a config on its own? Maybe even tie it with the running queries,
	// waiting for them to finish
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	api.logger.Debug("stopping called")

	err := api.srv.Shutdown(ctx)
	if err != nil {
		api.logger.Error("error when stopping", "error", err)
	}

	api.logger.Info("stopped")
}

func (api *GraviolaAPI) createRoutes() {
	router := chi.NewRouter()

	router.Use(httpmiddleware.NewLoggingMiddleware(api.logger))
	router.Use(httpmiddleware.NewMetricsMiddleware(api.metricRegistry))
	router.Use(middleware.Recoverer)

	router.Get("/metrics", promhttp.HandlerFor(api.metricRegistry, promhttp.HandlerOpts{Registry: api.metricRegistry}).ServeHTTP)
	router.Get("/healthy", alwaysSuccessfulHandler)
	router.Get("/ready", alwaysSuccessfulHandler)

	subRouter := route.New()
	subRouter = subRouter.WithPrefix("/api/v1")
	api.prometheusNativeAPI.Register(subRouter)
	router.Handle("/*", subRouter)

	api.router = router
	api.srv = &http.Server{Addr: fmt.Sprintf(":%d", api.conf.Port), Handler: router}
}

func alwaysSuccessfulHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
