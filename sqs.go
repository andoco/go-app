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
	endpoint     string
	receiveQueue string
	handler      MsgHandler
	queue        *Queue
	wg           *sync.WaitGroup
}

type SQSWorkerConfig struct {
	Endpoint     string
	ReceiveQueue string
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
		wg:           a.wg,
		endpoint:     config.Endpoint,
		receiveQueue: config.ReceiveQueue,
		handler:      handler,
	}

	a.sqsWorkers = append(a.sqsWorkers, s)
}

func (a *App) startSQSWorkers(ctx context.Context) {
	for _, ws := range a.sqsWorkers {
		setupQueue(ws)
		go workerLoop(ctx, ws)
		a.wg.Add(1)
	}
}

func setupQueue(state *sqsWorkerState) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	sess.Config.MergeIn(aws.NewConfig().WithEndpoint(state.endpoint))

	svc := sqs.New(sess)
	queue := NewQueue(NewQueueConfig(), svc)
	state.queue = queue
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
			messages, err := state.queue.Receive(ctx, state.receiveQueue, 1)
			if err != nil {
				fmt.Println(err)

				continue
			}

			for _, msg := range messages {
				if err := state.handler(msg); err != nil {
					fmt.Println(err)
					continue
				}

				if err = state.queue.Delete(ctx, msg, state.receiveQueue); err != nil {
					fmt.Println(err)
					continue
				}
			}
		}
	}
}
