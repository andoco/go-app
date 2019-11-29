package app

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type sqsWorkerState struct {
	endpoint        string
	receiveQueue    string
	deadLetterQueue string
	handler         MsgHandler
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
		endpoint:        config.Endpoint,
		receiveQueue:    config.ReceiveQueue,
		deadLetterQueue: config.DeadLetterQueue,
		handler:         handler,
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
	sess.Config.MergeIn(aws.NewConfig().WithEndpoint(state.endpoint))

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
				panic(err)
			}

			if _, err = svc.DeleteMessage(newDeleteMessageInput(state, msg)); err != nil {
				panic(err)
			}
		}
	}
}

type MsgAction int

const (
	MsgActionDelete MsgAction = iota
	MsgActionDeadLetter
)

func evalMsgAction(err error) MsgAction {
	if err == nil {
		return MsgActionDelete
	}

	return MsgActionDeadLetter
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
