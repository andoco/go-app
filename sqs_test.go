package app

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddSQS(t *testing.T) {
	os.Setenv("MY_APP_FOO_ENDPOINT", "test-endpoint")
	defer os.Unsetenv("MY_APP_FOO_ENDPOINT")
	os.Setenv("MY_APP_FOO_RECEIVEQUEUE", "test-queue")
	defer os.Unsetenv("MY_APP_FOO_RECEIVEQUEUE")

	app := NewApp(NewAppConfig("MyApp").Build())
	app.AddSQS("Foo", NewMsgRouter())

	assert.Len(t, app.sqsWorkers, 1)
	assert.Equal(t, "test-endpoint", app.sqsWorkers[0].endpoint)
	assert.Equal(t, "test-queue", app.sqsWorkers[0].receiveQueue)
	assert.Equal(t, "msgType", app.sqsWorkers[0].msgTypeKey)
	assert.NotNil(t, app.sqsWorkers[0].handler)
}

func TestAddSQSWithConfig(t *testing.T) {
	app := NewApp(NewAppConfig("MyApp").Build())
	c := &SQSWorkerConfig{
		Endpoint:     "test-endpoint",
		ReceiveQueue: "test-queue",
		MsgTypeKey:   "msgType",
	}
	app.AddSQSWithConfig(c, NewMsgRouter())

	assert.Len(t, app.sqsWorkers, 1)
	assert.Equal(t, "test-endpoint", app.sqsWorkers[0].endpoint)
	assert.Equal(t, "test-queue", app.sqsWorkers[0].receiveQueue)
	assert.Equal(t, "msgType", app.sqsWorkers[0].msgTypeKey)
	assert.NotNil(t, app.sqsWorkers[0].handler)
}

func TestNewSQSWorkerConfig(t *testing.T) {
	c := NewSQSWorkerConfig()

	assert.Equal(t, "msgType", c.MsgTypeKey)
}

func TestNewMessageContext(t *testing.T) {
	msg := &sqs.Message{}
	msg.SetMessageAttributes(map[string]*sqs.MessageAttributeValue{"msgType": &sqs.MessageAttributeValue{StringValue: aws.String("foo")}})

	msgCtx := newMessageContext(msg, "msgType", zerolog.New(os.Stderr))

	require.NotNil(t, msgCtx, "message context nil")
	assert.NotNil(t, msgCtx.Msg, "message not set")
	assert.Equal(t, "foo", *msgCtx.MsgType, "wrong msgType")
}
