package messages

import (
	"bytes"
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

var mockHash = types.Hash{
	0xbd, 0x79, 0x1d, 0x4a, 0xf9, 0x64, 0x8f, 0xc3, 0x7f, 0x94, 0xeb, 0x36, 0x53, 0x19, 0xf6, 0xd0,
	0xa9, 0x78, 0x9f, 0x9c, 0x22, 0x47, 0x2c, 0xa7, 0xa6, 0x12, 0xa9, 0xca, 0x4, 0x13, 0xc1, 0x4,
}

func TestEncodeMessage2(t *testing.T) {
	response := Message{
		MessageType: ResponseMessageType,
		PayLoad: &ResponseMessage{
			Response: &Response{
				Account:   mockAccounts[0],
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

func TestBroadcastPeersFilter(t *testing.T) {
	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, fmt.Errorf("resolve error")
	})
	BroadcastPeersFilter(nil, ProposalMessageType, mockHash, mockAccounts, mockAccounts[1])

	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, nil
	})
	monkey.Patch(net.DialTCP, func(string, *net.TCPAddr, *net.TCPAddr) (*net.TCPConn, error) {
		return nil, fmt.Errorf("dail error")
	})
	BroadcastPeersFilter(nil, ProposalMessageType, mockHash, mockAccounts, mockAccounts[1])

	var c net.TCPConn
	monkey.Patch(net.DialTCP, func(string, *net.TCPAddr, *net.TCPAddr) (*net.TCPConn, error) {
		return &c, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(&c), "Write", func(*net.TCPConn, []byte) (int, error) {
		return 0, nil
	})
	BroadcastPeersFilter(nil, ProposalMessageType, mockHash, mockAccounts, mockAccounts[1])

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

func mockBlocks(num int) []*types.Block {
	blocks := make([]*types.Block, 0)
	for index := 0; index < num; index++ {
		block := &types.Block{
			Header: &types.Header{
				Height: uint64(index),
			},
		}
		blocks = append(blocks, block)
	}
	return blocks
}

func TestReadMessage(t *testing.T) {
	request := Message{
		MessageType: RequestMessageType,
		PayLoad: &RequestMessage{
			Request: &Request{
				Account:   mockAccounts[0],
				Timestamp: time.Now().Unix(),
				Payload:   mockBlocks(1)[0],
			},
		},
	}
	msgRaw, err := EncodeMessage(request)
	assert.NotNil(t, msgRaw)
	assert.Nil(t, err)
	r := bytes.NewReader(msgRaw)
	message, err := ReadMessage(r)
	assert.Nil(t, err)
	assert.Equal(t, request, message)

	response := Message{
		MessageType: ProposalMessageType,
		PayLoad: &ProposalMessage{
			Proposal: &Proposal{
				Account:   mockAccounts[0],
				Timestamp: time.Now().Unix(),
				Signature: mockHash[:],
				Payload:   mockBlocks(1)[0],
			},
		},
	}
	msgRaw, err = EncodeMessage(response)
	assert.NotNil(t, msgRaw)
	assert.Nil(t, err)
	r = bytes.NewReader(msgRaw)
	message, err = ReadMessage(r)
	assert.Nil(t, err)
	assert.Equal(t, response, message)

	response = Message{
		MessageType: ResponseMessageType,
		PayLoad: &ResponseMessage{
			Response: &Response{
				Account:     mockAccounts[0],
				Timestamp:   time.Now().Unix(),
				Digest:      mockHash,
				Signature:   mockHash[:],
				BlockHeight: uint64(1),
			},
		},
	}
	msgRaw, err = EncodeMessage(response)
	assert.NotNil(t, msgRaw)
	assert.Nil(t, err)
	r = bytes.NewReader(msgRaw)
	message, err = ReadMessage(r)
	assert.Nil(t, err)
	assert.Equal(t, response, message)

	commit := Message{
		MessageType: CommitMessageType,
		PayLoad: &CommitMessage{
			Commit: &Commit{
				Account:    mockAccounts[0],
				Timestamp:  time.Now().Unix(),
				BlockHash:  mockHash,
				Digest:     mockHash,
				Signatures: [][]byte{mockHash[:], mockHash[:]},
				Result:     false,
			},
		},
	}
	msgRaw, err = EncodeMessage(commit)
	assert.NotNil(t, msgRaw)
	assert.Nil(t, err)
	r = bytes.NewReader(msgRaw)
	message, err = ReadMessage(r)
	assert.Nil(t, err)
	assert.Equal(t, commit, message)

	syncBlockReq := Message{
		MessageType: SyncBlockReqMessageType,
		PayLoad: &SyncBlockReqMessage{
			SyncBlockReq: &SyncBlockReq{
				Account:    mockAccounts[0],
				Timestamp:  time.Now().Unix(),
				BlockStart: uint64(1),
				BlockEnd:   uint64(2),
			},
		},
	}
	msgRaw, err = EncodeMessage(syncBlockReq)
	assert.NotNil(t, msgRaw)
	assert.Nil(t, err)
	r = bytes.NewReader(msgRaw)
	message, err = ReadMessage(r)
	assert.Nil(t, err)
	assert.Equal(t, syncBlockReq, message)

	syncBlockResp := Message{
		MessageType: SyncBlockRespMessageType,
		PayLoad: &SyncBlockRespMessage{
			SyncBlockResp: &SyncBlockResp{
				Blocks: mockBlocks(10),
			},
		},
	}
	msgRaw, err = EncodeMessage(syncBlockResp)
	assert.NotNil(t, msgRaw)
	assert.Nil(t, err)
	r = bytes.NewReader(msgRaw)
	message, err = ReadMessage(r)
	assert.Nil(t, err)
	assert.Equal(t, syncBlockResp, message)

	viewChangeReq := Message{
		MessageType: ViewChangeMessageReqType,
		PayLoad: &ViewChangeReqMessage{
			ViewChange: &ViewChangeReq{
				Account:   mockAccounts[0],
				Nodes:     mockAccounts,
				Timestamp: time.Now().Unix(),
				ViewNum:   uint64(1),
			},
		},
	}
	msgRaw, err = EncodeMessage(viewChangeReq)
	assert.NotNil(t, msgRaw)
	assert.Nil(t, err)
	r = bytes.NewReader(msgRaw)
	message, err = ReadMessage(r)
	assert.Nil(t, err)
	assert.Equal(t, viewChangeReq, message)

	onlineRequest := Message{
		MessageType: OnlineRequestType,
		PayLoad: &OnlineRequestMessage{
			OnlineRequest: &OnlineRequest{
				Account:     mockAccounts[0],
				Timestamp:   time.Now().Unix(),
				BlockHeight: uint64(1),
			},
		},
	}
	msgRaw, err = EncodeMessage(onlineRequest)
	assert.NotNil(t, msgRaw)
	assert.Nil(t, err)
	r = bytes.NewReader(msgRaw)
	message, err = ReadMessage(r)
	assert.Nil(t, err)
	assert.Equal(t, onlineRequest, message)

	onlineResponse := Message{
		MessageType: OnlineResponseType,
		PayLoad: &OnlineResponseMessage{
			OnlineResponse: &OnlineResponse{
				Account:     mockAccounts[0],
				Timestamp:   time.Now().Unix(),
				BlockHeight: uint64(1),
				Nodes:       mockAccounts,
				ViewNum:     uint64(1),
				Master:      mockAccounts[1],
			},
		},
	}
	msgRaw, err = EncodeMessage(onlineResponse)
	assert.NotNil(t, msgRaw)
	assert.Nil(t, err)
	r = bytes.NewReader(msgRaw)
	message, err = ReadMessage(r)
	assert.Nil(t, err)
	assert.Equal(t, onlineResponse, message)
}
