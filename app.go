package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

type App struct {
	name        string
	httpServers []*httpState
	sqsWorkers  []*sqsWorkerState
	wg          *sync.WaitGroup
	cancel      context.CancelFunc
	logger      zerolog.Logger
}

type AppConfig struct {
	Prometheus PrometheusConfig
}

type PrometheusConfig struct {
	Enabled bool
	Path    string `default:"/metrics"`
	Port    int    `default:"9090"`
}

type httpState struct {
	httpHandler http.Handler
	httpPort    int
	httpServer  *http.Server
}

func NewApp(name string) *App {
	logger := zerolog.New(os.Stderr).With().Str("appName", name).Logger()
	app := &App{name: name, wg: &sync.WaitGroup{}, logger: logger}

	appCfg := &AppConfig{}
	if err := app.ReadConfig(appCfg); err != nil {
		panic(err)
	}

	if appCfg.Prometheus.Enabled {
		app.AddPrometheus(appCfg.Prometheus.Path, appCfg.Prometheus.Port)
	}

	return app
}

func (a App) ReadConfig(c interface{}, name ...string) error {
	name = append([]string{a.name}, name...)
	return ReadEnvConfig(BuildEnvConfigName(name...), c)
}

func (a *App) AddPrometheus(path string, port int) {
	promMux := http.NewServeMux()
	promMux.Handle(path, promhttp.Handler())
	a.AddHttp(promMux, port)
}

func (a *App) AddHttp(handler http.Handler, port int) {
	s := &httpState{
		httpHandler: handler,
		httpPort:    port,
	}

	a.httpServers = append(a.httpServers, s)
}

func (a *App) Start() {
	a.logger.Debug().Msg("Starting app")
	ctx := context.Background()
	ctx2, cancel := context.WithCancel(ctx)

	a.cancel = cancel

	for _, s := range a.httpServers {
		s.httpServer = newHttpServer(s.httpHandler, s.httpPort)
		a.runListenAndServe(s)
		a.wg.Add(1)
	}

	a.startSQSWorkers(ctx2)

	a.registerStopOnSigTerm()
	a.wg.Wait()
}

func (a App) Stop() {
	a.logger.Debug().Msg("Stopping app")

	for _, s := range a.httpServers {
		if err := s.httpServer.Shutdown(context.TODO()); err != nil {
			panic(err)
		}
	}

	a.cancel()
}

func (a App) registerStopOnSigTerm() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		<-c
		a.Stop()
	}()
}

func newHttpServer(handler http.Handler, port int) *http.Server {
	return &http.Server{
		Handler: handler,
		Addr:    fmt.Sprintf(":%d", port),
	}
}

func (a *App) runListenAndServe(s *httpState) {
	go func() {
		defer a.wg.Done()

		if err := s.httpServer.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				panic(fmt.Errorf("server did not exit gracefully: %w", err))
			}
		}
		a.logger.Debug().Msg("HTTP server shutdown")
	}()
}
