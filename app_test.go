package app

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
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
	defer os.Unsetenv("MYAPP_FOO")

	app := NewApp("MyApp")
	app.ReadConfig(c)

	assert.Equal(t, "Foo", c.Foo)
}

func TestAddPrometheus(t *testing.T) {
	app := NewApp("MyApp")
	app.AddPrometheus("/metrics", 9090)

	assert.Len(t, app.httpServers, 1)
}

func TestAutoAddPrometheus(t *testing.T) {
	os.Setenv("MYAPP_PROMETHEUS_ENABLED", "true")
	defer os.Unsetenv("MYAPP_PROMETHEUS_ENABLED")
	os.Setenv("MYAPP_PROMETHEUS_PORT", "9999")
	defer os.Unsetenv("MYAPP_PROMETHEUS_PORT")

	app := NewApp("MyApp")

	require.Len(t, app.httpServers, 1)
	assert.Equal(t, 9999, app.httpServers[0].httpPort)
}

func TestNewLogger(t *testing.T) {
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
			logger := newLogger("MyApp", tc.env)
			assert.Equal(t, tc.outLogLevel, logger.GetLevel())
		})
	}
}
