package bft

import (
	"encoding/json"
	"fmt"
	"github.com/DSiSc/blockchain"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/policy/bft/messages"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/validator/tools/account"
	"github.com/DSiSc/validator/worker"
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

var mockHash = types.Hash{
	0xbd, 0x79, 0x1d, 0x4a, 0xf9, 0x64, 0x8f, 0xc3, 0x7f, 0x94, 0xeb, 0x36, 0x53, 0x19, 0xf6, 0xd0,
	0xa9, 0x78, 0x9f, 0x9c, 0x22, 0x47, 0x2c, 0xa7, 0xa6, 0x12, 0xa9, 0xca, 0x4, 0x13, 0xc1, 0x4,
}

func TestNewBFTCore(t *testing.T) {
	sigChannel := make(chan messages.SignatureSet)
	bft := NewBFTCore(mockAccounts[0], sigChannel)
	assert.NotNil(t, bft)
	assert.Equal(t, mockAccounts[0], bft.local)
}

func TestBftCore_ProcessEvent(t *testing.T) {
	sigChannel := make(chan messages.SignatureSet)
	bft := NewBFTCore(mockAccounts[0], sigChannel)
	assert.NotNil(t, bft)
	id := mockAccounts[0].Extension.Id
	err := bft.ProcessEvent(nil)
	assert.Nil(t, err)

	var b *blockchain.BlockChain
	monkey.Patch(blockchain.NewBlockChainByBlockHash, func(types.Hash) (*blockchain.BlockChain, error) {
		return b, nil
	})
	var w *worker.Worker
	monkey.PatchInstanceMethod(reflect.TypeOf(w), "VerifyBlock", func(*worker.Worker) error {
		return fmt.Errorf("verify block failed")
	})

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
		Payload: &types.Block{
			Header: &types.Header{
				Height:        0,
				PrevBlockHash: mockHash,
			},
		},
	}
	bft.master = id + 1
	err = bft.ProcessEvent(mock_proposal)
	assert.Nil(t, err)

	bft.master = id
	mockResponse := &messages.Response{
		Account:   mockAccounts[0],
		Timestamp: time.Now().Unix(),
		Digest:    mockHash,
		Signature: []byte{0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68},
	}
	go func() {
		err = bft.ProcessEvent(mockResponse)
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
	monkey.Unpatch(blockchain.NewBlockChainByBlockHash)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(w), "VerifyBlock")
}

func TestBftCore_Start(t *testing.T) {
	sigChannel := make(chan messages.SignatureSet)
	bft := NewBFTCore(mockAccounts[0], sigChannel)
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
	bft := NewBFTCore(mockAccounts[0], sigChannel)
	id := mockAccounts[0].Extension.Id
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
	bft := NewBFTCore(mockAccounts[0], sigChannel)
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
	bft := NewBFTCore(mockAccounts[0], sigChannel)
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
	bft := NewBFTCore(mockAccounts[0], sigChannel)
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

	// verify failed: Get NewBlockChainByBlockHash failed
	bft.local.Extension.Id = mockAccounts[0].Extension.Id + 1
	bft.receiveProposal(proposal)

	var b *blockchain.BlockChain
	monkey.Patch(blockchain.NewBlockChainByBlockHash, func(types.Hash) (*blockchain.BlockChain, error) {
		return b, nil
	})
	var w *worker.Worker
	monkey.PatchInstanceMethod(reflect.TypeOf(w), "VerifyBlock", func(*worker.Worker) error {
		return fmt.Errorf("verify block failed")
	})
	bft.receiveProposal(proposal)

	monkey.PatchInstanceMethod(reflect.TypeOf(w), "VerifyBlock", func(*worker.Worker) error {
		return nil
	})

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
	monkey.Unpatch(blockchain.NewBlockChainByBlockHash)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(w), "VerifyBlock")
}

func TestBftCore_receiveResponse(t *testing.T) {
	sigChannel := make(chan messages.SignatureSet)
	bft := NewBFTCore(mockAccounts[0], sigChannel)
	assert.NotNil(t, bft)
	bft.peers = mockAccounts
	id := mockAccounts[0].Extension.Id
	// master receive response
	response := &messages.Response{
		Account:   mockAccounts[0],
		Timestamp: time.Now().Unix(),
		Digest:    mockHash,
		Signature: []byte{0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68},
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
