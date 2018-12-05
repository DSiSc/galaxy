package messages

import (
	"github.com/DSiSc/craft/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncodeMessage(t *testing.T) {
	request := &RequestMessage{
		Request: &Request{
			Timestamp: int64(0),
			Payload: &types.Block{
				Header: &types.Header{
					Height: uint64(1),
				},
			},
		},
	}
	msg := Msg{
		MsgType: RequestMsgType,
		PayLoad: request,
	}
	rawData, err := EncodeMessage(msg)
	assert.Nil(t, err)
	assert.NotNil(t, rawData)

	ttt, err := decodeMessage(RequestMsgType, rawData[12:])
	assert.Nil(t, err)
	assert.Equal(t, ttt, msg)
}
