package app

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApp(t *testing.T) {
	app := NewApp("MyApp")

	assert.NotNil(t, app)
	assert.Nil(t, app.httpServers)
	assert.Nil(t, app.sqsWorkers)
	assert.Equal(t, "MyApp", app.name, "wrong app name")
}

func TestReadConfig(t *testing.T) {
	c := &struct {
		Foo string
	}{}

	os.Setenv("MYAPP_FOO", "Foo")

	app := NewApp("MyApp")
	app.ReadConfig(c)

	assert.Equal(t, "Foo", c.Foo)
}

func TestAddHttp(t *testing.T) {
	app := NewApp("MyApp")
	handler := http.NewServeMux()
	app.AddHttp(handler, 8080)

	assert.Len(t, app.httpServers, 1)
	assert.Equal(t, handler, app.httpServers[0].httpHandler, "Handler not added")
	assert.Equal(t, 8080, app.httpServers[0].httpPort, "Wrong port")
}

func TestAddPrometheus(t *testing.T) {
	app := NewApp("MyApp")
	app.AddPrometheus("/metrics", 9090)

	assert.Len(t, app.httpServers, 1)
}

func TestAutoAddPrometheus(t *testing.T) {
	os.Setenv("MYAPP_PROMETHEUS_ENABLED", "true")
	os.Setenv("MYAPP_PROMETHEUS_PORT", "9999")

	app := NewApp("MyApp")

	require.Len(t, app.httpServers, 1)
	assert.Equal(t, 9999, app.httpServers[0].httpPort)
}

func TestNewHTTPServer(t *testing.T) {
	handler := http.NewServeMux()
	server := newHttpServer(handler, 8081)

	assert.Equal(t, handler, server.Handler, "Handler not set")
	assert.Equal(t, ":8081", server.Addr, "Addr not set")
}
