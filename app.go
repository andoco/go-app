package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
)

type App struct {
	httpHandler http.Handler
	httpPort    int
	httpServer  *http.Server
}

func NewApp() *App {
	return &App{}
}

func (a *App) AddHttp(handler http.Handler, port int) {
	a.httpHandler = handler
	a.httpPort = port
}

func (a *App) Start() {
	a.httpServer = newHttpServer(a.httpHandler, a.httpPort)

	a.registerStopOnSigTerm()

	if err := a.httpServer.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Errorf("server did not exit gracefully: %w", err))
		}
	}
	fmt.Println("HTTP servers shutdown gracefully")
}

func (a App) Stop() {
	if err := a.httpServer.Shutdown(context.TODO()); err != nil {
		panic(err)
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
