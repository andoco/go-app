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

func TestDelete(t *testing.T) {
	testCases := []struct {
		name      string
		clientErr error
		err       string
	}{
		{name: "no error"},
		{name: "cancellation error", clientErr: awserr.New(request.CanceledErrorCode, "test-error", nil)},
		{name: "other error", clientErr: awserr.New("error-while-deleting", "test-error", nil), err: "error-while-deleting"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := &mockSQSClient{err: tc.clientErr}
			qconf := NewQueueConfig("msgType")
			queue := NewQueue(qconf, mockSvc)
			msg := &sqs.Message{MessageId: aws.String("test-message-id"), ReceiptHandle: aws.String("test-receipt-handle")}
			err := queue.Delete(context.TODO(), msg, "test-queue")

			if tc.err == "" {
				require.NoError(t, err)
				require.NotNil(t, mockSvc.deleteInput)
				assert.Equal(t, "test-queue", *mockSvc.deleteInput.QueueUrl)
				assert.Equal(t, "test-receipt-handle", *mockSvc.deleteInput.ReceiptHandle)
			} else {
				require.Error(t, err)
				assert.Regexp(t, tc.err, err.Error())
			}
		})
	}
}

func TestSend(t *testing.T) {
	for _, tc := range []struct {
		name      string
		clientErr error
		err       string
	}{
		{name: "no error"},
		{name: "cancellation error", clientErr: awserr.New(request.CanceledErrorCode, "test-error", nil)},
		{name: "other error", clientErr: awserr.New("error-while-sending", "test-error", nil), err: "error-while-sending"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := &mockSQSClient{err: tc.clientErr}
			qconf := NewQueueConfig("msgType")
			queue := NewQueue(qconf, mockSvc)
			err := queue.Send(context.TODO(), "test-body", "test-queue")

			if tc.err == "" {
				require.NoError(t, err)
				require.NotNil(t, mockSvc.sendInput)
				assert.Equal(t, "test-queue", *mockSvc.sendInput.QueueUrl)
				assert.Equal(t, "test-body", *mockSvc.sendInput.MessageBody)
			} else {
				require.Error(t, err)
				assert.Regexp(t, tc.err, err.Error())
			}
		})
	}
}

type mockSQSClient struct {
	sqsiface.SQSAPI
	err         error
	output      *sqs.ReceiveMessageOutput
	deleteInput *sqs.DeleteMessageInput
	sendInput   *sqs.SendMessageInput
}

func (m *mockSQSClient) ReceiveMessageWithContext(aws.Context, *sqs.ReceiveMessageInput, ...request.Option) (*sqs.ReceiveMessageOutput, error) {
	if m.output != nil {
		return m.output, nil
	}
	return nil, m.err
}

func (m *mockSQSClient) DeleteMessageWithContext(ctx aws.Context, input *sqs.DeleteMessageInput, options ...request.Option) (*sqs.DeleteMessageOutput, error) {
	m.deleteInput = input
	return nil, m.err
}

func (m *mockSQSClient) SendMessageWithContext(ctx aws.Context, input *sqs.SendMessageInput, options ...request.Option) (*sqs.SendMessageOutput, error) {
	m.sendInput = input
	return nil, m.err
}
