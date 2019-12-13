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
	config      *AppConfig
	httpServers []*httpState
	sqsWorkers  []*sqsWorkerState
	tasks       []*taskState
	wg          *sync.WaitGroup
	cancel      context.CancelFunc
	logger      zerolog.Logger
	Metrics     *Metrics
}

type AppConfig struct {
	Name       string
	Env        string `default:"dev"`
	Prometheus PrometheusConfig
}

func NewAppConfig(name string) *AppConfig {
	return &AppConfig{Name: name}
}

func (c AppConfig) WithMetrics(prefix string) *AppConfig {
	c.Prometheus = PrometheusConfig{
		Prefix: prefix,
	}

	return &c
}

// NewApp creates a new App. name is expected to be in upper camelcase format.
func NewApp(config *AppConfig) *App {
	logger := newLogger(config.Name)

	if !validateAppName(config.Name) {
		logger.Fatal().Msg("Invalid app name")
	}

	app := &App{
		config: config,
		wg:     &sync.WaitGroup{},
		logger: logger,
	}

	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			logger.Fatal().Err(err).Msg("Error loading .env file")
		}
	}

	if err := app.ReadConfig(app.config); err != nil {
		logger.Fatal().Err(err).Msg("Error reading core app configuration")
	}

	app.logger = loggerForEnv(logger, app.config.Env)
	app.Metrics = NewMetrics(app.config.Prometheus)

	if app.config.Prometheus.Enabled {
		app.AddPrometheus(app.config.Prometheus.Path, app.config.Prometheus.Port)
	}

	return app
}

func (a App) ReadConfig(c interface{}, name ...string) error {
	splitAppName := splitUpperCamelCase(a.config.Name)
	path := append(splitAppName, name...)
	return ReadEnvConfig(c, path...)
}

func (a *App) AddPrometheus(path string, port int) {
	promMux := http.NewServeMux()
	promMux.Handle(path, promhttp.HandlerFor(a.Metrics.registry, promhttp.HandlerOpts{}))
	a.AddHttp(promMux, port)
}

func (a *App) Start() {
	a.logger.Debug().Msg("Starting app")
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	a.cancel = cancel

	a.startHttpServers(ctx)
	a.startSQSWorkers(ctx)
	a.startTasks(ctx)

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
