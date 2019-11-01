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
	return &App{httpPort: 8081}
}

func (a *App) AddHttp(handler http.Handler) {
	a.httpHandler = handler
}

func (a *App) Start() {
	server := &http.Server{
		Handler: a.httpHandler,
		Addr:    fmt.Sprintf(":%d", a.httpPort),
	}

	a.httpServer = server

	a.registerStopOnSigTerm()

	if err := server.ListenAndServe(); err != nil {
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
