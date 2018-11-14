package bft

import (
	"encoding/json"
	"fmt"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/policy/bft/messages"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/validator/tools/account"
	"github.com/stretchr/testify/assert"
	"net"
	"reflect"
	"testing"
	"time"
)

var id uint64 = 0

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

func TestNewBFTCore(t *testing.T) {
	sigChannel := make(chan messages.SignatureSet)
	bft := NewBFTCore(id, sigChannel)
	assert.NotNil(t, bft)
	assert.Equal(t, id, bft.id)
}

func TestBftCore_ProcessEvent(t *testing.T) {
	sigChannel := make(chan messages.SignatureSet)
	bft := NewBFTCore(id, sigChannel)
	assert.NotNil(t, bft)
	err := bft.ProcessEvent(nil)
	assert.Nil(t, err)

	bft.peers = mockAccounts
	var mockSignature = [][]byte{{0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68}}
	var mock_request = &messages.Request{
		Timestamp: time.Now().Unix(),
		Payload: &types.Block{
			Header: &types.Header{
				SigData: mockSignature,
			},
		},
	}
	monkey.Patch(json.Marshal, func(v interface{}) ([]byte, error) {
		return nil, nil
	})
	monkey.Patch(net.DialTCP, func(string, *net.TCPAddr, *net.TCPAddr) (*net.TCPConn, error) {
		return nil, nil
	})
	var c net.TCPConn
	monkey.Patch(net.DialTCP, func(string, *net.TCPAddr, *net.TCPAddr) (*net.TCPConn, error) {
		return &c, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(&c), "Write", func(*net.TCPConn, []byte) (int, error) {
		return 0, nil
	})
	err = bft.ProcessEvent(mock_request)
	assert.Nil(t, err)

	var mock_proposal = &messages.Proposal{
		Timestamp: time.Now().Unix(),
		Payload:   nil,
	}
	bft.master = id + 1
	err = bft.ProcessEvent(mock_proposal)
	assert.Nil(t, err)

	bft.master = id
	mock_response := &messages.Response{
		Id:        id,
		Timestamp: time.Now().Unix(),
		Payload:   nil,
		Signature: []byte{0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68},
	}
	go func() {
		err = bft.ProcessEvent(mock_response)
		assert.Nil(t, err)
	}()
	ch := <-bft.result
	assert.NotNil(t, ch)
	var exceptSig = make([][]byte, 0)
	exceptSig = append(exceptSig, []byte{0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68})
	assert.Equal(t, messages.SignatureSet(exceptSig), ch)
	monkey.Unpatch(net.ResolveTCPAddr)
	monkey.Unpatch(net.DialTCP)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(&c), "Write")
}

func TestBftCore_Start(t *testing.T) {
	sigChannel := make(chan messages.SignatureSet)
	bft := NewBFTCore(id, sigChannel)
	assert.NotNil(t, bft)
	var account = account.Account{
		Extension: account.AccountExtension{
			Url: "127.0.0.1:8080",
		},
	}
	go bft.Start(account)
	time.Sleep(1 * time.Second)
}

func TestBftCore_receiveRequest(t *testing.T) {
	sigChannel := make(chan messages.SignatureSet)
	bft := NewBFTCore(id, sigChannel)
	assert.NotNil(t, bft)
	bft.peers = mockAccounts
	// only master process request
	request := &messages.Request{
		Timestamp: 1535414400,
		Payload: &types.Block{
			Header: &types.Header{
				Height:  0,
				SigData: make([][]byte, 0),
			},
		},
	}
	bft.master = id + 1
	bft.receiveRequest(request)
	// absence of signature
	bft.master = id
	bft.receiveRequest(request)
	//  marshal failed
	var fakeSignature = []byte{
		0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33,
		0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d,
	}
	request.Payload.Header.SigData = append(request.Payload.Header.SigData, fakeSignature)
	monkey.Patch(json.Marshal, func(interface{}) ([]byte, error) {
		return nil, fmt.Errorf("marshal proposal msg failed")
	})
	bft.receiveRequest(request)
	monkey.Patch(json.Marshal, func(interface{}) ([]byte, error) {
		return nil, nil
	})
	monkey.Patch(net.DialTCP, func(string, *net.TCPAddr, *net.TCPAddr) (*net.TCPConn, error) {
		return nil, nil
	})
	var c net.TCPConn
	monkey.Patch(net.DialTCP, func(string, *net.TCPAddr, *net.TCPAddr) (*net.TCPConn, error) {
		return &c, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(&c), "Write", func(*net.TCPConn, []byte) (int, error) {
		return 0, nil
	})
	bft.receiveRequest(request)
	monkey.Unpatch(net.ResolveTCPAddr)
	monkey.Unpatch(net.DialTCP)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(&c), "Write")
}

func TestNewBFTCore_broadcast(t *testing.T) {
	sigChannel := make(chan messages.SignatureSet)
	bft := NewBFTCore(id, sigChannel)
	assert.NotNil(t, bft)
	bft.peers = mockAccounts
	// resolve error
	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, fmt.Errorf("resolve error")
	})
	bft.broadcast(nil)

	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, nil
	})
	monkey.Patch(net.DialTCP, func(string, *net.TCPAddr, *net.TCPAddr) (*net.TCPConn, error) {
		return nil, fmt.Errorf("dail error")
	})
	bft.broadcast(nil)

	var c net.TCPConn
	monkey.Patch(net.DialTCP, func(string, *net.TCPAddr, *net.TCPAddr) (*net.TCPConn, error) {
		return &c, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(&c), "Write", func(*net.TCPConn, []byte) (int, error) {
		return 0, nil
	})
	bft.broadcast(nil)
	monkey.Unpatch(net.ResolveTCPAddr)
	monkey.Unpatch(net.DialTCP)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(&c), "Write")
}

func TestBftCore_unicast(t *testing.T) {
	sigChannel := make(chan messages.SignatureSet)
	bft := NewBFTCore(id, sigChannel)
	assert.NotNil(t, bft)
	bft.peers = mockAccounts
	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, fmt.Errorf("resolve error")
	})
	err := bft.unicast(bft.peers[1], nil)
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
	err = bft.unicast(bft.peers[1], nil)
	assert.Nil(t, err)
	monkey.Unpatch(net.ResolveTCPAddr)
	monkey.Unpatch(net.DialTCP)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(&c), "Write")
}

func TestBftCore_receiveProposal(t *testing.T) {
	sigChannel := make(chan messages.SignatureSet)
	bft := NewBFTCore(id, sigChannel)
	assert.NotNil(t, bft)
	bft.peers = mockAccounts
	// master receive proposal
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
	monkey.Patch(json.Marshal, func(interface{}) ([]byte, error) {
		return nil, fmt.Errorf("marshal proposal msg failed")
	})
	bft.receiveProposal(proposal)

	monkey.Patch(json.Marshal, func(interface{}) ([]byte, error) {
		return nil, nil
	})
	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, fmt.Errorf("resolve error")
	})
	bft.receiveProposal(proposal)
	monkey.Unpatch(net.ResolveTCPAddr)
	monkey.Unpatch(json.Marshal)
}

func TestBftCore_receiveResponse(t *testing.T) {
	sigChannel := make(chan messages.SignatureSet)
	bft := NewBFTCore(id, sigChannel)
	assert.NotNil(t, bft)
	bft.peers = mockAccounts
	// master receive response
	response := &messages.Response{
		Timestamp: 1535414400,
		Payload: &types.Block{
			Header: &types.Header{
				Height:  0,
				SigData: make([][]byte, 0),
			},
		},
	}
	bft.master = id + 1
	bft.receiveResponse(response)

	bft.master = id
	bft.tolerance = 3
	// signature not exist
	var fakeSignature = []byte{
		0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33,
		0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d,
	}
	response.Signature = fakeSignature
	bft.receiveResponse(response)

	// signature exist, while not consistence
	var fakeSignature1 = []byte{
		0x34, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33,
		0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d,
	}
	response.Signature = fakeSignature1
	bft.receiveResponse(response)

	// all conditions are met
	bft.tolerance = 1
	response.Signature = fakeSignature

	go func() {
		bft.receiveResponse(response)
	}()
	ch := <-bft.result
	assert.NotNil(t, ch)
	var exceptSig = make([][]byte, 0)
	exceptSig = append(exceptSig, fakeSignature)
	assert.Equal(t, messages.SignatureSet(exceptSig), ch)
}
