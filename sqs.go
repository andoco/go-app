package app

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type sqsWorkerState struct {
	endpoint     string
	receiveQueue string
	msgTypeKey   string
	handler      MsgHandler
	queue        *Queue
	wg           *sync.WaitGroup
	logger       zerolog.Logger
	metrics      *sqsMetrics
}

type sqsMetrics struct {
	msgReceived         *prometheus.CounterVec
	msgProcessed        *prometheus.CounterVec
	msgProcessedFailure *prometheus.CounterVec
	msgDeleted          *prometheus.CounterVec
}

type SQSWorkerConfig struct {
	Endpoint     string
	ReceiveQueue string
	MsgTypeKey   string
}

func NewSQSWorkerConfig() *SQSWorkerConfig {
	return &SQSWorkerConfig{
		MsgTypeKey: "msgType",
	}
}

type MsgContext struct {
	Msg     *sqs.Message
	MsgType *string
	Logger  zerolog.Logger
}

type MsgHandler interface {
	Process(msg *MsgContext) error
}

type MsgHandlerFunc func(msg *MsgContext) error

func (f MsgHandlerFunc) Process(msg *MsgContext) error {
	return f(msg)
}

func (a *App) AddSQS(prefix string, handler MsgHandler) {
	c := NewSQSWorkerConfig()
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
		msgTypeKey:   config.MsgTypeKey,
		handler:      handler,
		logger:       a.logger.With().Str("queue", config.ReceiveQueue).Logger(),
	}

	s.metrics = &sqsMetrics{
		msgReceived:         a.NewCounterVec("sqs_msg_received_total", "The total number of SQS messages received", []string{"app", "queue"}),
		msgProcessed:        a.NewCounterVec("sqs_msg_processed_total", "The total number of SQS messages processed", []string{"app", "queue"}),
		msgProcessedFailure: a.NewCounterVec("sqs_msg_processed_failure_total", "The total number of SQS messages that failed to be processed", []string{"app", "queue"}),
		msgDeleted:          a.NewCounterVec("sqs_msg_deleted_total", "The total number of SQS messages deleted", []string{"app", "queue"}),
	}

	a.sqsWorkers = append(a.sqsWorkers, s)
}

func (a *App) startSQSWorkers(ctx context.Context) {
	for _, ws := range a.sqsWorkers {
		setupQueue(ws)
		ws.logger.Debug().Msg("Starting queue worker")
		go workerLoop(ctx, a.name, ws)
		a.wg.Add(1)
	}
}

func setupQueue(state *sqsWorkerState) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	sess.Config.MergeIn(aws.NewConfig().WithEndpoint(state.endpoint))

	svc := sqs.New(sess)
	queue := NewQueue(NewQueueConfig(state.msgTypeKey), svc)
	state.queue = queue
}

func newMessageContext(msg *sqs.Message, msgTypeKey string, logger zerolog.Logger) *MsgContext {
	var msgType *string
	if msgTypeAttrib, ok := msg.MessageAttributes[msgTypeKey]; ok {
		msgType = msgTypeAttrib.StringValue
	}

	return &MsgContext{
		Msg:     msg,
		MsgType: msgType,
		Logger:  logger,
	}
}

func workerLoop(ctx context.Context, appName string, state *sqsWorkerState) {
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

			state.metrics.msgReceived.With(prometheus.Labels{"app": appName, "queue": state.receiveQueue}).Add(float64(len(messages)))

			for _, msg := range messages {
				logger := state.logger.With().Str("messageId", *msg.MessageId).Logger()

				msgCtx := newMessageContext(msg, state.msgTypeKey, logger)

				if err := state.handler.Process(msgCtx); err != nil {
					logger.Error().Err(err).Msg("Failed to handle message")
					state.metrics.msgProcessedFailure.With(prometheus.Labels{"app": appName, "queue": state.receiveQueue}).Inc()
					continue
				}

				state.metrics.msgProcessed.With(prometheus.Labels{"app": appName, "queue": state.receiveQueue}).Inc()

				if err = state.queue.Delete(ctx, msg, state.receiveQueue); err != nil {
					logger.Error().Err(err).Msg("Failed to delete message")
					continue
				}

				state.metrics.msgDeleted.With(prometheus.Labels{"app": appName, "queue": state.receiveQueue}).Inc()
			}
		}
	}
}
