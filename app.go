package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
)

type App struct {
	name        string
	httpServers []*httpState
	sqsWorkers  []*sqsWorkerState
	wg          *sync.WaitGroup
	cancel      context.CancelFunc
}

type httpState struct {
	httpHandler http.Handler
	httpPort    int
	httpServer  *http.Server
}

func NewApp(name string) *App {
	return &App{name: name, wg: &sync.WaitGroup{}}
}

func (a App) ReadConfig(c interface{}) error {
	return ReadEnvConfig(a.name, c)
}

func (a *App) AddHttp(handler http.Handler, port int) {
	s := &httpState{
		httpHandler: handler,
		httpPort:    port,
	}

	a.httpServers = append(a.httpServers, s)
}

func (a *App) Start() {
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
	for _, s := range a.httpServers {
		if err := s.httpServer.Shutdown(context.TODO()); err != nil {
			panic(err)
		}
	}

	a.cancel()

	fmt.Println("App stopped")
}

func (a App) registerStopOnSigTerm() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		<-c
		fmt.Println()
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
		fmt.Println("HTTP server shutdown gracefully")
	}()
}
