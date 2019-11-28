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
	wg          sync.WaitGroup
}

type httpState struct {
	httpHandler http.Handler
	httpPort    int
	httpServer  *http.Server
}

func NewApp(name string) *App {
	return &App{name: name}
}

func (a *App) AddHttp(handler http.Handler, port int) {
	s := &httpState{
		httpHandler: handler,
		httpPort:    port,
	}

	a.httpServers = append(a.httpServers, s)
}

func (a *App) Start() {
	for _, s := range a.httpServers {
		s.httpServer = newHttpServer(s.httpHandler, s.httpPort)
		a.registerStopOnSigTerm()
		a.runListenAndServe(s)
		a.wg.Add(1)
	}

	a.wg.Wait()
}

func (a App) Stop() {
	for _, s := range a.httpServers {
		if err := s.httpServer.Shutdown(context.TODO()); err != nil {
			panic(err)
		}
	}
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
		fmt.Println("HTTP servers shutdown gracefully")
	}()
}
