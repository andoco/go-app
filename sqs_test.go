package app

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
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
