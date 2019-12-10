package app

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewMsgRouter(t *testing.T) {
	mr := NewMsgRouter()
	require.NotNil(t, mr)
	assert.NotNil(t, mr.routes)
}

func TestMsgRouterHandle(t *testing.T) {
	mr := NewMsgRouter()
	handler := NewMsgRouter()
	mr.Handle("foo", handler)

	require.Len(t, mr.routes, 1)
	assert.Contains(t, mr.routes, "foo")
	h, ok := mr.routes["foo"]
	require.True(t, ok)
	assert.Equal(t, handler, h)
}

func TestMsgRouterProcess(t *testing.T) {
	msgType := "foo"
	otherMsgType := "bar"

	testCases := []struct {
		name       string
		msgType    *string
		outHandled bool
		outErr     error
	}{
		{name: "no msgType", msgType: nil, outHandled: false, outErr: NoMsgTypeErr},
		{name: "matching msgType", msgType: &msgType, outHandled: true, outErr: nil},
		{name: "non-matching msgType", msgType: &otherMsgType, outHandled: false, outErr: nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mr := NewMsgRouter()

			var handled bool
			mr.HandleFunc(msgType, func(msg *MsgContext) error {
				handled = true
				return nil
			})

			msg := &sqs.Message{}
			if tc.msgType != nil {
				msg.SetMessageAttributes(map[string]*sqs.MessageAttributeValue{"msgType": &sqs.MessageAttributeValue{StringValue: tc.msgType}})
			}

			msgCtx := newMessageContext(msg, "msgType", zerolog.New(nil))

			err := mr.Process(msgCtx)

			if tc.outErr == nil {
				require.Nil(t, err, "Process returned error")
			} else {
				require.Equal(t, tc.outErr, err)
			}

			assert.Equal(t, tc.outHandled, handled, "message handled")
		})
	}
}
