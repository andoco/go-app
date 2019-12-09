package app

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	NoMsgTypeErr = errors.New("no msgType attribute found on message")

	msgRouted = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sf_go_app_sqs_msg_routed_total",
		Help: "The total number of SQS messages routed to a handler",
	}, []string{"msgType"})
)

// MsgRouter handles routing messages to handlers based
// on the message's msgType.
type MsgRouter struct {
	routes map[string]MsgHandler
}

func NewMsgRouter() *MsgRouter {
	return &MsgRouter{routes: make(map[string]MsgHandler)}
}

// Handle registers handler to receive messages of msgType.
func (r *MsgRouter) Handle(msgType string, handler MsgHandler) {
	r.routes[msgType] = handler
}

// Handle registers handler to receive messages of msgType.
func (r *MsgRouter) HandleFunc(msgType string, handler MsgHandlerFunc) {
	r.routes[msgType] = handler
}

// Dispatch will pass msg to the registered handler for the
// message's msgType.
func (r MsgRouter) Process(msg *MsgContext) error {
	msg.Logger.Debug().Msg("Routing message")

	if msg.MsgType == "" {
		return NoMsgTypeErr
	}

	msg.Logger = msg.Logger.With().Str("msgType", msg.MsgType).Logger()
	msg.Logger.Debug().Msg("Found msgType")

	h, ok := r.routes[msg.MsgType]
	if !ok {
		return nil
	}

	msg.Logger.Debug().Msg("Processing message")

	msgRouted.With(prometheus.Labels{"msgType": msg.MsgType}).Inc()

	return h.Process(msg)
}
