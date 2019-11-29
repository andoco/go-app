package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type sqsWorkerState struct {
	endpoint        string
	receiveQueue    string
	deadLetterQueue string
	handler         MsgHandler
	receiver        messageReceiver
	sender          messageSender
	deleter         messageDeleter
	wg              *sync.WaitGroup
}

type SQSWorkerConfig struct {
	Endpoint        string
	ReceiveQueue    string
	DeadLetterQueue string
}

type MsgHandler func(msg *sqs.Message) error

func (a *App) AddSQS(prefix string, handler MsgHandler) {
	c := &SQSWorkerConfig{}
	if err := ReadEnvConfig(fmt.Sprintf("%s_%s", a.name, prefix), c); err != nil {
		panic(err)
	}

	a.AddSQSWithConfig(c, handler)
}

func (a *App) AddSQSWithConfig(config *SQSWorkerConfig, handler MsgHandler) {
	s := &sqsWorkerState{
		wg:              a.wg,
		endpoint:        config.Endpoint,
		receiveQueue:    config.ReceiveQueue,
		deadLetterQueue: config.DeadLetterQueue,
		handler:         handler,
	}

	a.sqsWorkers = append(a.sqsWorkers, s)
}

func (a *App) startSQSWorkers(ctx context.Context) {
	for _, ws := range a.sqsWorkers {
		setupMessageFuncs(ws)
		go workerLoop(ctx, ws)
		a.wg.Add(1)
	}
}

type messageReceiver func(ctx context.Context) ([]*sqs.Message, error)
type messageSender func(ctx context.Context, queue string, body string) error
type messageDeleter func(ctx context.Context, msg *sqs.Message) error

func setupMessageFuncs(state *sqsWorkerState) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	sess.Config.MergeIn(aws.NewConfig().WithEndpoint(state.endpoint))

	svc := sqs.New(sess)

	state.receiver = func(ctx context.Context) ([]*sqs.Message, error) {
		input := newReceiveMessageInput(state)
		output, err := svc.ReceiveMessageWithContext(ctx, input)
		if err != nil {
			return nil, err
		}

		return output.Messages, nil
	}

	state.deleter = func(ctx context.Context, msg *sqs.Message) error {
		input := newDeleteMessageInput(state, msg)
		_, err := svc.DeleteMessageWithContext(ctx, input)
		if err != nil {
			return err
		}
		return nil
	}

	state.sender = func(ctx context.Context, queue, body string) error {
		input := newSendMessageInput(state, queue, body)
		_, err := svc.SendMessageWithContext(ctx, input)
		if err != nil {
			return err
		}
		return nil
	}
}

func workerLoop(ctx context.Context, state *sqsWorkerState) {
	defer state.wg.Done()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Exiting SQS worker loop")
			return
		default:
			fmt.Println("Receiving messages")
			messages, err := state.receiver(ctx)
			if err != nil {
				continue
			}

			for _, msg := range messages {
				if err := state.handler(msg); err != nil {
					if err := state.sender(ctx, state.deadLetterQueue, *msg.Body); err != nil {
						continue
					}
					if err := state.deleter(ctx, msg); err != nil {
						continue
					}
				}

				if err = state.deleter(ctx, msg); err != nil {
					continue
				}
			}
		}
	}
}

func newReceiveMessageInput(state *sqsWorkerState) *sqs.ReceiveMessageInput {
	return &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(state.receiveQueue),
		MaxNumberOfMessages: aws.Int64(1),
		WaitTimeSeconds:     aws.Int64(10),
	}
}

func newDeleteMessageInput(state *sqsWorkerState, msg *sqs.Message) *sqs.DeleteMessageInput {
	return &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(state.receiveQueue),
		ReceiptHandle: msg.ReceiptHandle,
	}
}

func newSendMessageInput(state *sqsWorkerState, queue string, body string) *sqs.SendMessageInput {
	return &sqs.SendMessageInput{
		QueueUrl:    aws.String(queue),
		MessageBody: aws.String(body),
	}
}
