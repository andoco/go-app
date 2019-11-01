package app

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddHttp(t *testing.T) {
	app := NewApp()
	handler := http.NewServeMux()
	app.AddHttp(handler)

	assert.Equal(t, handler, app.httpHandler, "Handler not added")
}
