package app

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddHttp(t *testing.T) {
	app := NewApp(NewAppConfig("MyApp"))
	handler := http.NewServeMux()
	app.AddHttp(handler, 8080)

	require.Len(t, app.httpServers, 1)
	assert.Equal(t, handler, app.httpServers[0].httpHandler, "Handler not added")
	assert.Equal(t, 8080, app.httpServers[0].httpPort, "Wrong port")
}

func TestNewHTTPServer(t *testing.T) {
	handler := http.NewServeMux()
	server := newHttpServer(handler, 8081)

	assert.Equal(t, handler, server.Handler, "Handler not set")
	assert.Equal(t, ":8081", server.Addr, "Addr not set")
}
