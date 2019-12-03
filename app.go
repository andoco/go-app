package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/joho/godotenv"

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
	Env        string `default:"dev"`
	Prometheus PrometheusConfig
}

type PrometheusConfig struct {
	Enabled bool
	Path    string `default:"/metrics"`
	Port    int    `default:"9090"`
}

func NewApp(name string) *App {
	app := &App{
		name:   name,
		wg:     &sync.WaitGroup{},
		logger: newLogger(name),
	}

	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			app.logger.Fatal().Err(err).Msg("Error loading .env file")
		}
	}

	appCfg := &AppConfig{}
	if err := app.ReadConfig(appCfg); err != nil {
		app.logger.Fatal().Err(err).Msg("Error reading core app configuration")
	}

	app.logger = loggerForEnv(app.logger, appCfg.Env)

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

func (a *App) Start() {
	a.logger.Debug().Msg("Starting app")
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	a.cancel = cancel

	a.startHttpServers(ctx)
	a.startSQSWorkers(ctx)

	a.registerStopOnSigTerm()
	a.wg.Wait()
}

func (a App) Stop() {
	a.logger.Debug().Msg("Stopping app")

	a.stopHttpServers(context.TODO())

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

func newLogger(name string) zerolog.Logger {
	return zerolog.New(os.Stderr).With().Str("appName", name).Logger()
}

func logLevelForEnv(env string) zerolog.Level {
	var logLevel zerolog.Level

	switch env {
	case "dev":
		logLevel = zerolog.DebugLevel
	default:
		logLevel = zerolog.WarnLevel
	}

	return logLevel
}

func loggerForEnv(logger zerolog.Logger, env string) zerolog.Logger {
	logLevel := logLevelForEnv(env)
	return logger.Level(logLevel)
}
