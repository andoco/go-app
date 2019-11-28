package app

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewApp(t *testing.T) {
	app := NewApp()

	assert.NotNil(t, app)
}

func TestAddHttp(t *testing.T) {
	app := NewApp()
	handler := http.NewServeMux()
	app.AddHttp(handler)

	assert.Equal(t, handler, app.httpHandler, "Handler not added")
}

func TestNewHTTPServer(t *testing.T) {
	handler := http.NewServeMux()
	server := newHttpServer(handler, 8081)

	assert.Equal(t, handler, server.Handler, "Handler not set")
	assert.Equal(t, ":8081", server.Addr, "Addr not set")
}
