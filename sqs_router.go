package app

import (
	"errors"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/rs/zerolog"
)

var (
	NoMsgTypeErr = errors.New("no msgType attribute found on message")
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
func (r MsgRouter) Process(msg *sqs.Message, logger zerolog.Logger) error {
	logger.Debug().Msg("Routing message")

	msgTypeAttr, ok := msg.MessageAttributes["msgType"];
	if !ok {
		return NoMsgTypeErr
	}

	msgType := *msgTypeAttr.StringValue

	logger = logger.With().Str("msgType", msgType).Logger()
	logger.Debug().Msg("Found msgType")

	h, ok := r.routes[msgType]
	if !ok {
		return nil
	}

	logger.Debug().Msg("Processing message")

	return h.Process(msg, logger)
}
