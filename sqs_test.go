package app

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stretchr/testify/assert"
)

func TestAddSQS(t *testing.T) {
	os.Setenv("MYAPP_FOO_ENDPOINT", "test-endpoint")
	os.Setenv("MYAPP_FOO_RECEIVEQUEUE", "test-queue")
	os.Setenv("MYAPP_FOO_DEADLETTERQUEUE", "dead-letter-queue")

	app := NewApp("MyApp")
	app.AddSQS("Foo", func(_ *sqs.Message) error { return nil })

	assert.Len(t, app.sqsWorkers, 1)
	assert.Equal(t, "test-endpoint", app.sqsWorkers[0].endpoint)
	assert.Equal(t, "test-queue", app.sqsWorkers[0].receiveQueue)
	assert.Equal(t, "dead-letter-queue", app.sqsWorkers[0].deadLetterQueue)
	assert.NotNil(t, app.sqsWorkers[0].handler)
}

func TestAddSQSWithConfig(t *testing.T) {
	app := NewApp("MyApp")
	c := &SQSWorkerConfig{
		Endpoint:        "test-endpoint",
		ReceiveQueue:    "test-queue",
		DeadLetterQueue: "dead-letter-queue",
	}
	app.AddSQSWithConfig(c, func(_ *sqs.Message) error { return nil })

	assert.Len(t, app.sqsWorkers, 1)
	assert.Equal(t, "test-endpoint", app.sqsWorkers[0].endpoint)
	assert.Equal(t, "test-queue", app.sqsWorkers[0].receiveQueue)
	assert.Equal(t, "dead-letter-queue", app.sqsWorkers[0].deadLetterQueue)
	assert.NotNil(t, app.sqsWorkers[0].handler)
}

func TestNewReceiveMessageInput(t *testing.T) {
	state := &sqsWorkerState{
		receiveQueue: "test-queue",
	}

	rmi := newReceiveMessageInput(state)

	assert.NotNil(t, rmi)
	assert.Equal(t, aws.String("test-queue"), rmi.QueueUrl)
	assert.Equal(t, aws.Int64(10), rmi.WaitTimeSeconds)
	assert.Equal(t, aws.Int64(1), rmi.MaxNumberOfMessages)
}

func TestNewDeleteMessageInput(t *testing.T) {
	state := &sqsWorkerState{
		receiveQueue: "test-queue",
	}

	msg := &sqs.Message{ReceiptHandle: aws.String("test-handle")}

	dmi := newDeleteMessageInput(state, msg)

	assert.NotNil(t, dmi)
	assert.Equal(t, aws.String("test-queue"), dmi.QueueUrl)
	assert.Equal(t, aws.String("test-handle"), dmi.ReceiptHandle)
}
