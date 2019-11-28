package app

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stretchr/testify/assert"
)

func TestAddSQS(t *testing.T) {
	app := NewApp("MyApp")
	app.AddSQS("test-queue", func(_ *sqs.Message) error { return nil })

	assert.Len(t, app.sqsWorkers, 1)
	assert.Equal(t, "test-queue", app.sqsWorkers[0].receiveQueue)
	assert.NotNil(t, app.sqsWorkers[0].handler)
}
