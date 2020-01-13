package app

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestNewReceiveMessageInput(t *testing.T) {
	rmi := newReceiveMessageInput("test-queue", 1, 10, "msgType")

	assert.NotNil(t, rmi)
	assert.Equal(t, aws.String("test-queue"), rmi.QueueUrl)
	assert.Equal(t, aws.Int64(10), rmi.WaitTimeSeconds)
	assert.Equal(t, aws.Int64(1), rmi.MaxNumberOfMessages)
	assert.Contains(t, rmi.MessageAttributeNames, aws.String("msgType"))
}

func TestNewDeleteMessageInput(t *testing.T) {
	msg := &sqs.Message{ReceiptHandle: aws.String("test-handle")}

	dmi := newDeleteMessageInput("test-queue", msg)

	assert.NotNil(t, dmi)
	assert.Equal(t, aws.String("test-queue"), dmi.QueueUrl)
	assert.Equal(t, aws.String("test-handle"), dmi.ReceiptHandle)
}

func TestReceiveWhenCancelledReturnsNoMessages(t *testing.T) {
	mockSvc := &mockSQSClient{err: awserr.New(request.CanceledErrorCode, "", nil)}
	qconf := NewQueueConfig("msgType")
	queue := NewQueue(qconf, mockSvc)
	msgs, err := queue.Receive(context.TODO(), "foo", 1)
	assert.NoError(t, err)
	assert.Empty(t, msgs)
}

func TestReceiveWhenNonCancelledErrorReturnsError(t *testing.T) {
	mockSvc := &mockSQSClient{err: awserr.New("some error", "", nil)}
	qconf := NewQueueConfig("msgType")
	queue := NewQueue(qconf, mockSvc)
	msgs, err := queue.Receive(context.TODO(), "foo", 1)
	assert.Error(t, err)
	assert.Nil(t, msgs)
}

type mockSQSClient struct {
	sqsiface.SQSAPI
	err error
}

func (m *mockSQSClient) ReceiveMessageWithContext(aws.Context, *sqs.ReceiveMessageInput, ...request.Option) (*sqs.ReceiveMessageOutput, error) {
	return nil, m.err
}
