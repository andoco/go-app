package app

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddSQS(t *testing.T) {
	os.Setenv("MYAPP_FOO_ENDPOINT", "test-endpoint")
	defer os.Unsetenv("MYAPP_FOO_ENDPOINT")
	os.Setenv("MYAPP_FOO_RECEIVEQUEUE", "test-queue")
	defer os.Unsetenv("MYAPP_FOO_RECEIVEQUEUE")

	app := NewApp("MyApp")
	app.AddSQS("Foo", NewMsgRouter())

	assert.Len(t, app.sqsWorkers, 1)
	assert.Equal(t, "test-endpoint", app.sqsWorkers[0].endpoint)
	assert.Equal(t, "test-queue", app.sqsWorkers[0].receiveQueue)
	assert.NotNil(t, app.sqsWorkers[0].handler)
}

func TestAddSQSWithConfig(t *testing.T) {
	app := NewApp("MyApp")
	c := &SQSWorkerConfig{
		Endpoint:     "test-endpoint",
		ReceiveQueue: "test-queue",
	}
	app.AddSQSWithConfig(c, NewMsgRouter())

	assert.Len(t, app.sqsWorkers, 1)
	assert.Equal(t, "test-endpoint", app.sqsWorkers[0].endpoint)
	assert.Equal(t, "test-queue", app.sqsWorkers[0].receiveQueue)
	assert.NotNil(t, app.sqsWorkers[0].handler)
}
