package app

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestReceive(t *testing.T) {
	output := &sqs.ReceiveMessageOutput{Messages: []*sqs.Message{{Body: aws.String("test-body")}}}
	mockSvc := &mockSQSClient{output: output}
	qconf := NewQueueConfig("msgType")
	queue := NewQueue(qconf, mockSvc)
	msgs, err := queue.Receive(context.TODO(), "foo", 1)
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	assert.Equal(t, "test-body", *msgs[0].Body)
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
	err    error
	output *sqs.ReceiveMessageOutput
}

func (m *mockSQSClient) ReceiveMessageWithContext(aws.Context, *sqs.ReceiveMessageInput, ...request.Option) (*sqs.ReceiveMessageOutput, error) {
	if m.output != nil {
		return m.output, nil
	}
	return nil, m.err
}
