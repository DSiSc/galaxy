package bft

import (
	"fmt"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/policy/bft/messages"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/validator/tools/account"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"net"
	"reflect"
	"testing"
	"time"
)

var id uint64 = 0

func TestNewBFTCore(t *testing.T) {
	receive := NewBFTCore(id, id, nil)
	assert.NotNil(t, receive)
	bftcore := receive.(*bftCore)
	assert.Equal(t, id, bftcore.id)
}

func TestBftCore_ProcessEvent(t *testing.T) {
	receive := NewBFTCore(id, id, nil)
	assert.NotNil(t, receive)
	bftcore := receive.(*bftCore)
	err := bftcore.ProcessEvent(nil)
	assert.Nil(t, err)

	var mock_request = messages.Request{
		Timestamp: time.Now().Unix(),
		Payload:   nil,
	}

	err = bftcore.ProcessEvent(mock_request)
	assert.Nil(t, err)

	var mock_proposal = messages.Proposal{
		Timestamp: time.Now().Unix(),
		Payload:   nil,
	}
	err = bftcore.ProcessEvent(mock_proposal)
	assert.Nil(t, err)
}

func TestBftCore_Start(t *testing.T) {
	receive := NewBFTCore(id, id, nil)
	assert.NotNil(t, receive)
	bftCore := receive.(*bftCore)
	var account = account.Account{
		Extension: account.AccountExtension{
			Url: "127.0.0.1:8080",
		},
	}
	go bftCore.Start(account)
	time.Sleep(2 * time.Second)
}

func TestBftCore_receiveRequest(t *testing.T) {
	receive := NewBFTCore(id, 1, nil)
	assert.NotNil(t, receive)
	bft := receive.(*bftCore)
	request := &messages.Request{
		Timestamp: 1535414400,
		Payload: &types.Block{
			Header: &types.Header{
				Height: 0,
			},
		},
	}
	bft.receiveRequest(request)
	bft.id = id + 1
	monkey.Patch(proto.Marshal, func(proto.Message) ([]byte, error) {
		return nil, fmt.Errorf("marshal proposal msg failed")
	})
	bft.receiveRequest(request)
	monkey.Patch(proto.Marshal, func(proto.Message) ([]byte, error) {
		return nil, nil
	})
	bft.receiveRequest(request)
}

func mockAccounts() []account.Account {
	account_0 := account.Account{
		Address: types.Address{0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68,
			0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  0,
			Url: "172.0.0.1:8080",
		},
	}

	account_1 := account.Account{
		Address: types.Address{0x34, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68,
			0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  1,
			Url: "172.0.0.1:8081",
		},
	}
	return []account.Account{account_0, account_1}
}

func TestNewBFTCore_broadcast(t *testing.T) {
	receive := NewBFTCore(id, id, nil)
	assert.NotNil(t, receive)
	bftCore := receive.(*bftCore)
	bftCore.peers = mockAccounts()
	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, fmt.Errorf("resolve error")
	})
	bftCore.broadcast(nil)

	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, nil
	})
	monkey.Patch(net.DialTCP, func(string, *net.TCPAddr, *net.TCPAddr) (*net.TCPConn, error) {
		return nil, fmt.Errorf("dail error")
	})
	bftCore.broadcast(nil)

	var c net.TCPConn
	monkey.Patch(net.DialTCP, func(string, *net.TCPAddr, *net.TCPAddr) (*net.TCPConn, error) {
		return &c, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(&c), "Write", func(*net.TCPConn, []byte) (int, error) {
		return 0, nil
	})
	bftCore.broadcast(nil)
}

func TestBftCore_unicast(t *testing.T) {
	receive := NewBFTCore(id, id, nil)
	assert.NotNil(t, receive)
	bftCore := receive.(*bftCore)
	bftCore.peers = mockAccounts()
	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, fmt.Errorf("resolve error")
	})
	err := bftCore.unicast(bftCore.peers[1], nil)
	assert.Equal(t, fmt.Errorf("resolve error"), err)
	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, nil
	})
	var c net.TCPConn
	monkey.Patch(net.DialTCP, func(string, *net.TCPAddr, *net.TCPAddr) (*net.TCPConn, error) {
		return &c, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(&c), "Write", func(*net.TCPConn, []byte) (int, error) {
		return 0, nil
	})
	err = bftCore.unicast(bftCore.peers[1], nil)
	assert.Nil(t, err)
}

func TestBftCore_receiveProposal(t *testing.T) {
	receive := NewBFTCore(id, id, nil)
	assert.NotNil(t, receive)
	bft := receive.(*bftCore)
	proposal := &messages.Proposal{
		Timestamp: 1535414400,
		Payload: &types.Block{
			Header: &types.Header{
				Height: 0,
			},
		},
	}
	bft.receiveProposal(proposal)
	bft.id = id + 1

	monkey.Patch(proto.Marshal, func(proto.Message) ([]byte, error) {
		return nil, fmt.Errorf("marshal proposal msg failed")
	})
	bft.receiveProposal(proposal)
	bft.peers = mockAccounts()

	monkey.Patch(proto.Marshal, func(proto.Message) ([]byte, error) {
		return nil, nil
	})
	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, fmt.Errorf("resolve error")
	})
	bft.receiveProposal(proposal)
}
