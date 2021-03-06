package app

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAppConfig(t *testing.T) {
	a := NewAppConfig("FooBar")

	assert.Equal(t, "FooBar", a.Name)
}

func TestNewApp(t *testing.T) {
	app := NewApp(NewAppConfig("MyApp").Build())

	assert.NotNil(t, app)
	assert.Nil(t, app.httpServers)
	assert.Nil(t, app.sqsWorkers)
	assert.Equal(t, "MyApp", app.config.Name, "wrong app name")
	assert.NotNil(t, app.Metrics)
}

func TestReadConfig(t *testing.T) {
	c := &struct {
		Foo string
	}{}

	os.Setenv("MY_APP_FOO", "Foo")
	defer os.Unsetenv("MY_APP_FOO")

	app := NewApp(NewAppConfig("MyApp").Build())
	app.ReadConfig(c)

	assert.Equal(t, "Foo", c.Foo)
}

func TestAddPrometheus(t *testing.T) {
	app := NewApp(NewAppConfig("MyApp").Build())
	app.AddPrometheus("/metrics", 9090)

	assert.Len(t, app.httpServers, 1)
}

func TestAutoAddPrometheus(t *testing.T) {
	os.Setenv("MY_APP_PROMETHEUS_ENABLED", "true")
	defer os.Unsetenv("MY_APP_PROMETHEUS_ENABLED")
	os.Setenv("MY_APP_PROMETHEUS_PORT", "9999")
	defer os.Unsetenv("MY_APP_PROMETHEUS_PORT")

	app := NewApp(NewAppConfig("MyApp").Build())

	require.Len(t, app.httpServers, 1)
	assert.Equal(t, 9999, app.httpServers[0].httpPort)
}

func TestLogLevelForEnv(t *testing.T) {
	testCases := []struct {
		name        string
		env         string
		outLogLevel zerolog.Level
	}{
		{"dev", "dev", zerolog.DebugLevel},
		{"staging", "staging", zerolog.WarnLevel},
		{"production", "production", zerolog.WarnLevel},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logLevel := logLevelForEnv(tc.env)
			assert.Equal(t, tc.outLogLevel, logLevel)
		})
	}
}
