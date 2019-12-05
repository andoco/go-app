package app

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/rs/zerolog"
)

type sqsWorkerState struct {
	endpoint     string
	receiveQueue string
	handler      MsgHandler
	queue        *Queue
	wg           *sync.WaitGroup
	logger       zerolog.Logger
}

type SQSWorkerConfig struct {
	Endpoint     string
	ReceiveQueue string
}

type MsgHandler func(msg *sqs.Message, logger zerolog.Logger) error

func (a *App) AddSQS(prefix string, handler MsgHandler) {
	c := &SQSWorkerConfig{}
	if err := ReadEnvConfig(BuildEnvConfigName(a.name, prefix), c); err != nil {
		a.logger.Fatal().Err(err).Str("prefix", prefix).Msg("Cannot read configuration")
	}

	a.AddSQSWithConfig(c, handler)
}

func (a *App) AddSQSWithConfig(config *SQSWorkerConfig, handler MsgHandler) {
	s := &sqsWorkerState{
		wg:           a.wg,
		endpoint:     config.Endpoint,
		receiveQueue: config.ReceiveQueue,
		handler:      handler,
		logger:       a.logger.With().Str("queue", config.ReceiveQueue).Logger(),
	}

	a.sqsWorkers = append(a.sqsWorkers, s)
}

func (a *App) startSQSWorkers(ctx context.Context) {
	for _, ws := range a.sqsWorkers {
		setupQueue(ws)
		ws.logger.Debug().Msg("Starting queue worker")
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
			state.logger.Debug().Msg("Exiting worker loop")
			return
		default:
			state.logger.Debug().Msg("Receiving messages")
			messages, err := state.queue.Receive(ctx, state.receiveQueue, 1)
			if err != nil {
				state.logger.Error().Err(err).Msg("Failed to receive message")
				continue
			}

			state.logger.Debug().Int("numMessages", len(messages)).Msg("Received messages")

			for _, msg := range messages {
				logger := state.logger.With().Str("messageId", *msg.MessageId).Logger()

				if err := state.handler(msg, logger); err != nil {
					logger.Error().Err(err).Msg("Failed to handle message")
					continue
				}

				if err = state.queue.Delete(ctx, msg, state.receiveQueue); err != nil {
					logger.Error().Err(err).Msg("Failed to delete message")
					continue
				}
			}
		}
	}
}
