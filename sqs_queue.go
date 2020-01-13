package app

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

// Receiver is an interface for receiving messages from a queue.
type Receiver interface {
	Receive(ctx context.Context, queue string, max int) ([]*sqs.Message, error)
}

// Deleter is an interface for deleting a message from a queue.
type Deleter interface {
	Delete(ctx context.Context, msg *sqs.Message, queue string) error
}

type Sender interface {
	Send(ctx context.Context, body string, queue string) error
}

func NewQueue(config *QueueConfig, svc sqsiface.SQSAPI) *Queue {
	return &Queue{config: config, svc: svc}
}

type Queue struct {
	config *QueueConfig
	svc    sqsiface.SQSAPI
}

func NewQueueConfig(msgTypeKey string) *QueueConfig {
	return &QueueConfig{
		WaitTime:   20,
		MsgTypeKey: msgTypeKey,
	}
}

type QueueConfig struct {
	WaitTime   int
	MsgTypeKey string
}

func (q Queue) Receive(ctx context.Context, queue string, max int) ([]*sqs.Message, error) {
	input := newReceiveMessageInput(queue, max, q.config.WaitTime, q.config.MsgTypeKey)

	output, err := q.svc.ReceiveMessageWithContext(ctx, input)
	if err != nil {
		if isAwsCancelledError(err) {
			return []*sqs.Message{}, nil
		}
		return nil, fmt.Errorf("receiving messages from %q: %w", queue, err)
	}

	return output.Messages, nil
}

func (q Queue) Delete(ctx context.Context, msg *sqs.Message, queue string) error {
	input := newDeleteMessageInput(queue, msg)

	_, err := q.svc.DeleteMessageWithContext(ctx, input)
	if err != nil {
		if isAwsCancelledError(err) {
			return nil
		}
		return fmt.Errorf("deleting message %q from %q: %w", *msg.MessageId, queue, err)
	}

	return nil
}

func (q Queue) Send(ctx context.Context, body string, queue string) error {
	input := newSendMessageInput(queue, body)

	_, err := q.svc.SendMessageWithContext(ctx, input)
	if err != nil {
		if isAwsCancelledError(err) {
			return nil
		}
		return fmt.Errorf("sending message to %q: %w", queue, err)
	}

	return nil
}

func isAwsCancelledError(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		return awsErr.Code() == request.CanceledErrorCode
	}
	return false
}

func newReceiveMessageInput(queue string, maxNumMessages int, waitTimeSeconds int, msgTypeKey string) *sqs.ReceiveMessageInput {
	input := &sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(queue),
		MaxNumberOfMessages:   aws.Int64(int64(maxNumMessages)),
		WaitTimeSeconds:       aws.Int64(int64(waitTimeSeconds)),
		MessageAttributeNames: aws.StringSlice([]string{msgTypeKey}),
	}

	return input
}

func newDeleteMessageInput(queue string, msg *sqs.Message) *sqs.DeleteMessageInput {
	return &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queue),
		ReceiptHandle: msg.ReceiptHandle,
	}
}

func newSendMessageInput(queue string, body string) *sqs.SendMessageInput {
	return &sqs.SendMessageInput{
		QueueUrl:    aws.String(queue),
		MessageBody: aws.String(body),
	}
}
