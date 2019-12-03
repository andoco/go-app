package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

type httpState struct {
	httpHandler http.Handler
	httpPort    int
	httpServer  *http.Server
}

func (a *App) AddHttp(handler http.Handler, port int) {
	s := &httpState{
		httpHandler: handler,
		httpPort:    port,
	}

	a.httpServers = append(a.httpServers, s)
}

func (a *App) startHttpServers(ctx context.Context) {
	for _, s := range a.httpServers {
		s.httpServer = newHttpServer(s.httpHandler, s.httpPort)
		a.runListenAndServe(s)
		a.wg.Add(1)
	}
}

func (a *App) stopHttpServers(ctx context.Context) {
	for _, s := range a.httpServers {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			panic(err)
		}
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

func newHttpServer(handler http.Handler, port int) *http.Server {
	return &http.Server{
		Handler: handler,
		Addr:    fmt.Sprintf(":%d", port),
	}
}
