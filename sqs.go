package app

import "github.com/aws/aws-sdk-go/service/sqs"

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
