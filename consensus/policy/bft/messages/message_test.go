package messages

import (
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/validator/tools/account"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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
	msg := Message{
		MessageType: RequestMessageType,
		PayLoad:     request,
	}
	rawData, err := EncodeMessage(msg)
	assert.Nil(t, err)
	assert.NotNil(t, rawData)

	ttt, err := DecodeMessage(RequestMessageType, rawData[12:])
	assert.Nil(t, err)
	assert.Equal(t, ttt, msg)
}

func TestEncodeMessage2(t *testing.T) {
	mockAccount := account.Account{
		Address: types.Address{0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68,
			0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  0,
			Url: "172.0.0.1:8080",
		},
	}
	mockHash := types.Hash{
		0xbd, 0x79, 0x1d, 0x4a, 0xf9, 0x64, 0x8f, 0xc3, 0x7f, 0x94, 0xeb, 0x36, 0x53, 0x19, 0xf6, 0xd0,
		0xa9, 0x78, 0x9f, 0x9c, 0x22, 0x47, 0x2c, 0xa7, 0xa6, 0x12, 0xa9, 0xca, 0x4, 0x13, 0xc1, 0x4,
	}
	response := Message{
		MessageType: ResponseMessageType,
		PayLoad: &ResponseMessage{
			Response: &Response{
				Account:   mockAccount,
				Timestamp: time.Now().Unix(),
				Digest:    mockHash,
				Signature: mockHash[:],
			},
		},
	}
	msgRaw, err := EncodeMessage(response)
	assert.NotNil(t, msgRaw)
	assert.Nil(t, err)
}
