package app

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type sqsWorkerState struct {
	receiveQueue string
	handler      MsgHandler
}

type MsgHandler func(msg *sqs.Message) error

func (a *App) AddSQS(queue string, handler MsgHandler) {
	s := &sqsWorkerState{
		receiveQueue: queue,
		handler:      handler,
	}

	a.sqsWorkers = append(a.sqsWorkers, s)
}

func (a *App) startSQSWorkers() {
	for _, ws := range a.sqsWorkers {
		go workerLoop(ws)
		a.wg.Add(1)
	}
}

func workerLoop(state *sqsWorkerState) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := sqs.New(sess)
	receiveMessageInput := newReceiveMessageInput(state)

	for {
		result, err := svc.ReceiveMessage(receiveMessageInput)

		if err != nil {
			panic(err)
		}

		for _, msg := range result.Messages {
			if err := state.handler(msg); err != nil {
				// TOOD: dead-letter
			}

			if _, err = svc.DeleteMessage(newDeleteMessageInput(state, msg)); err != nil {
				panic(err)
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
