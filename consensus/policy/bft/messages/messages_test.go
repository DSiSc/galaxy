package messages

import (
	"encoding/json"
	"fmt"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/validator/tools/account"
	"github.com/stretchr/testify/assert"
	"net"
	"reflect"
	"testing"
	"time"
)

var mockAccounts = []account.Account{
	account.Account{
		Address: types.Address{0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68,
			0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  0,
			Url: "172.0.0.1:8080",
		},
	},
	account.Account{
		Address: types.Address{0x34, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68,
			0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  1,
			Url: "172.0.0.1:8081"},
	},
	account.Account{
		Address: types.Address{0x35, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  2,
			Url: "172.0.0.1:8082",
		},
	},

	account.Account{
		Address: types.Address{0x36, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  3,
			Url: "172.0.0.1:8083",
		},
	},
}

func TestMessage_MessageType(t *testing.T) {
	assert.Equal(t, MessageType("RequestMessage"), RequestMessageType)
	assert.Equal(t, MessageType("ProposalMessage"), ProposalMessageType)
	assert.Equal(t, MessageType("ResponseMessage"), ResponseMessageType)
}

func mockMessage(messageType MessageType) *Message {

	var fakeSignature1 = []byte{
		0x34, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33,
		0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d,
	}
	payload := &types.Block{
		Header: &types.Header{
			Height: 1,
		},
	}

	if RequestMessageType == messageType {
		payload := &RequestMessage{
			Request: &Request{
				Timestamp: time.Now().Unix(),
				Payload:   payload,
			},
		}
		return &Message{
			MessageType: RequestMessageType,
			Payload:     payload,
		}
	}

	if ProposalMessageType == messageType {
		payload := &ProposalMessage{
			Proposal: &Proposal{
				Id:        0,
				Timestamp: time.Now().Unix(),
				Payload:   payload,
				Signature: fakeSignature1,
			},
		}
		return &Message{
			MessageType: ProposalMessageType,
			Payload:     payload,
		}
	}

	if ResponseMessageType == messageType {
		payload := &ResponseMessage{
			Response: &Response{
				Account:   mockAccounts[0],
				Timestamp: time.Now().Unix(),
				Signature: fakeSignature1,
			},
		}
		return &Message{
			MessageType: ResponseMessageType,
			Payload:     payload,
		}
	}
	return nil
}

func TestMessage_MarshalJSON(t *testing.T) {
	request := mockMessage(RequestMessageType)
	requestData, err := json.Marshal(request)
	assert.Nil(t, err)
	assert.NotNil(t, requestData)

	monkey.Patch(json.Marshal, func(v interface{}) ([]byte, error) {
		return nil, fmt.Errorf("marshal failed")
	})
	requestData, err = json.Marshal(request)
	assert.Nil(t, requestData)
	monkey.Unpatch(json.Marshal)
}

func TestMessage_UnmarshalJSON(t *testing.T) {
	requestMessage := mockMessage(RequestMessageType)
	requestData, err := json.Marshal(requestMessage)
	assert.Nil(t, err)
	assert.NotNil(t, requestData)
	var request Message
	err = json.Unmarshal(requestData, &request)
	assert.Nil(t, err)
	assert.Equal(t, requestMessage.MessageType, request.MessageType)
	assert.Equal(t, requestMessage.Payload, request.Payload)

	proposalMessage := mockMessage(ProposalMessageType)
	proposalData, err := json.Marshal(proposalMessage)
	assert.Nil(t, err)
	assert.NotNil(t, proposalData)
	var proposal Message
	err = json.Unmarshal(proposalData, &proposal)
	assert.Equal(t, proposalMessage.MessageType, proposal.MessageType)
	assert.Equal(t, proposalMessage.Payload, proposal.Payload)

	responseMessage := mockMessage(ResponseMessageType)
	responseData, err := json.Marshal(responseMessage)
	assert.Nil(t, err)
	assert.NotNil(t, responseData)
	var response Message
	err = json.Unmarshal(responseData, &response)
	assert.Equal(t, responseMessage.MessageType, response.MessageType)
	assert.Equal(t, responseMessage.Payload, response.Payload)

	responseMessage.MessageType = "unknown message type "
	responseData, err = json.Marshal(responseMessage)
	assert.Nil(t, err)
	assert.NotNil(t, responseData)
	err = json.Unmarshal(responseData, &response)
	assert.Equal(t, fmt.Errorf("not support marshal type"), err)
}

var mockHash = types.Hash{
	0xbd, 0x79, 0x1d, 0x4a, 0xf9, 0x64, 0x8f, 0xc3, 0x7f, 0x94, 0xeb, 0x36, 0x53, 0x19, 0xf6, 0xd0,
	0xa9, 0x78, 0x9f, 0x9c, 0x22, 0x47, 0x2c, 0xa7, 0xa6, 0x12, 0xa9, 0xca, 0x4, 0x13, 0xc1, 0x4,
}

func TestNewBFTCore_broadcast(t *testing.T) {
	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, fmt.Errorf("resolve error")
	})
	BroadcastPeers(nil, ProposalMessageType, mockHash, mockAccounts)

	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, nil
	})
	monkey.Patch(net.DialTCP, func(string, *net.TCPAddr, *net.TCPAddr) (*net.TCPConn, error) {
		return nil, fmt.Errorf("dail error")
	})
	BroadcastPeers(nil, ProposalMessageType, mockHash, mockAccounts)

	var c net.TCPConn
	monkey.Patch(net.DialTCP, func(string, *net.TCPAddr, *net.TCPAddr) (*net.TCPConn, error) {
		return &c, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(&c), "Write", func(*net.TCPConn, []byte) (int, error) {
		return 0, nil
	})
	BroadcastPeers(nil, ProposalMessageType, mockHash, mockAccounts)

	monkey.UnpatchAll()
}

func TestBftCore_unicast(t *testing.T) {
	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, fmt.Errorf("resolve error")
	})
	err := Unicast(mockAccounts[1], nil, ProposalMessageType, mockHash)
	assert.Equal(t, fmt.Errorf("resolve error"), err)

	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, nil
	})
	var c net.TCPConn
	monkey.Patch(net.DialTCP, func(string, *net.TCPAddr, *net.TCPAddr) (*net.TCPConn, error) {
		return &c, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(&c), "Write", func(*net.TCPConn, []byte) (int, error) {
		return 10, nil
	})
	err = Unicast(mockAccounts[1], nil, ProposalMessageType, mockHash)
	assert.NotNil(t, err)

	monkey.UnpatchAll()
}
