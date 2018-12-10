package fbft

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DSiSc/blockchain"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/messages"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/validator/tools/account"
	"github.com/DSiSc/validator/tools/signature"
	"github.com/DSiSc/validator/tools/signature/keypair"
	"github.com/DSiSc/validator/worker"
	"github.com/stretchr/testify/assert"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"
)

var events types.EventCenter

type Event struct {
	m           sync.RWMutex
	Subscribers map[types.EventType]map[types.Subscriber]types.EventFunc
}

func NewEvent() types.EventCenter {
	return &Event{
		Subscribers: make(map[types.EventType]map[types.Subscriber]types.EventFunc),
	}
}

//  adds a new subscriber to Event.
func (e *Event) Subscribe(eventType types.EventType, eventFunc types.EventFunc) types.Subscriber {
	e.m.Lock()
	defer e.m.Unlock()

	sub := make(chan interface{})
	_, ok := e.Subscribers[eventType]
	if !ok {
		e.Subscribers[eventType] = make(map[types.Subscriber]types.EventFunc)
	}
	e.Subscribers[eventType][sub] = eventFunc

	return sub
}

func (e *Event) UnSubscribe(eventType types.EventType, subscriber types.Subscriber) (err error) {
	e.m.Lock()
	defer e.m.Unlock()

	subEvent, ok := e.Subscribers[eventType]
	if !ok {
		err = errors.New("event type not exist")
		return
	}

	delete(subEvent, subscriber)
	close(subscriber)

	return
}

func (e *Event) Notify(eventType types.EventType, value interface{}) (err error) {

	e.m.RLock()
	defer e.m.RUnlock()

	subs, ok := e.Subscribers[eventType]
	if !ok {
		err = errors.New("event type not register")
		return
	}

	switch value.(type) {
	case error:
		log.Error("Receive errors is [%v].", value)
	}
	log.Info("Receive eventType is [%d].", eventType)

	for _, event := range subs {
		go e.NotifySubscriber(event, value)
	}
	return nil
}

func (e *Event) NotifySubscriber(eventFunc types.EventFunc, value interface{}) {
	if eventFunc == nil {
		return
	}

	// invoke subscriber event func
	eventFunc(value)

}

//Notify all event subscribers
func (e *Event) NotifyAll() (errs []error) {
	e.m.RLock()
	defer e.m.RUnlock()

	for eventType, _ := range e.Subscribers {
		if err := e.Notify(eventType, nil); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// unsubscribe all event and subscriber elegant
func (e *Event) UnSubscribeAll() {
	for eventtype, _ := range e.Subscribers {
		subs, ok := e.Subscribers[eventtype]
		if !ok {
			continue
		}
		for subscriber, _ := range subs {
			delete(subs, subscriber)
			close(subscriber)
		}
	}
	return
}

func TestNewBFTCore(t *testing.T) {
	bft := NewFBFTCore(mockAccounts[0], nil)
	assert.NotNil(t, bft)
	assert.Equal(t, mockAccounts[0], bft.nodes.local)
}

func TestBftCore_ProcessEvent(t *testing.T) {
	block := &types.Block{
		Header: &types.Header{
			Height:    uint64(1),
			MixDigest: mockHash,
		},
	}
	fbft := NewFBFTCore(mockAccounts[0], nil)
	fbft.timeoutTimer = time.NewTimer(30 * time.Second)
	fbft.consensusPlugin = common.NewConsensusPlugin()
	content := fbft.consensusPlugin.Add(mockHash, block)
	assert.NotNil(t, content)

	err := fbft.ProcessEvent(nil)
	assert.Equal(t, fmt.Errorf("not support type <nil>"), err)

	var mock_request = &messages.Request{
		Timestamp: time.Now().Unix(),
		Payload: &types.Block{
			Header: &types.Header{
				SigData: mockSignset[:1],
			},
		},
	}
	fbft.ProcessEvent(mock_request)

	fbft.nodes.master = mockAccounts[0]
	var b *blockchain.BlockChain
	monkey.Patch(blockchain.NewBlockChainByBlockHash, func(types.Hash) (*blockchain.BlockChain, error) {
		return b, nil
	})
	var w *worker.Worker
	monkey.PatchInstanceMethod(reflect.TypeOf(w), "VerifyBlock", func(*worker.Worker) error {
		return fmt.Errorf("verify block failed")
	})
	monkey.Patch(signature.Verify, func(keypair.PublicKey, []byte) (types.Address, error) {
		return mockAccounts[0].Address, nil
	})
	fbft.nodes.peers = mockAccounts
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
	fbft.ProcessEvent(mock_request)

	var mock_proposal = &messages.Proposal{
		Timestamp: time.Now().Unix(),
		Payload: &types.Block{
			Header: &types.Header{
				Height:        0,
				PrevBlockHash: mockHash,
			},
		},
	}
	fbft.nodes.master = mockAccounts[1]
	fbft.ProcessEvent(mock_proposal)
	monkey.Patch(signature.Verify, func(_ keypair.PublicKey, sign []byte) (types.Address, error) {
		var address types.Address
		if bytes.Equal(sign[:], mockSignset[0]) {
			address = mockAccounts[0].Address
		}
		if bytes.Equal(sign[:], mockSignset[1]) {
			address = mockAccounts[1].Address
		}
		if bytes.Equal(sign[:], mockSignset[2]) {
			address = mockAccounts[2].Address
		}
		if bytes.Equal(sign[:], mockSignset[3]) {
			address = mockAccounts[3].Address
		}
		return address, nil
	})

	fbft.nodes.master = mockAccounts[0]
	mockResponse := &messages.Response{
		Account:   mockAccounts[0],
		Timestamp: time.Now().Unix(),
		Digest:    mockHash,
		Signature: mockSignset[0],
	}
	ok := content.AddSignature(mockAccounts[0], mockSignset[0])
	assert.Equal(t, true, ok)
	ok = content.AddSignature(mockAccounts[1], mockSignset[1])
	assert.Equal(t, true, ok)
	ok = content.AddSignature(mockAccounts[2], mockSignset[2])
	assert.Equal(t, true, ok)
	fbft.tolerance = uint8((len(fbft.nodes.peers) - 1) / 3)
	go fbft.waitResponse(mockHash)
	fbft.ProcessEvent(mockResponse)
	ch := <-fbft.result
	assert.NotNil(t, ch)
	assert.Equal(t, 3, len(ch.Signatures))

	mockCommit := &messages.Commit{
		Account:    mockAccounts[0],
		Timestamp:  time.Now().Unix(),
		Digest:     mockHash,
		Signatures: mockSignset,
		BlockHash:  mockHash,
		Result:     true,
	}
	fbft.ProcessEvent(mockCommit)
	monkey.Unpatch(net.ResolveTCPAddr)
	monkey.Unpatch(net.DialTCP)
	monkey.Unpatch(signature.Verify)
	monkey.Unpatch(blockchain.NewBlockChainByBlockHash)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(&c), "Write")
	monkey.UnpatchInstanceMethod(reflect.TypeOf(w), "VerifyBlock")
}

func TestBftCore_Start(t *testing.T) {
	fbft := NewFBFTCore(mockAccounts[0], nil)
	assert.NotNil(t, fbft)
	var account = account.Account{
		Extension: account.AccountExtension{
			Url: "127.0.0.1:8080",
		},
	}
	commit := &messages.Commit{
		Account:    mockAccounts[0],
		Timestamp:  time.Now().Unix(),
		Digest:     mockHash,
		Signatures: mockSignset,
		BlockHash:  mockHash,
		Result:     true,
	}
	committed := messages.Message{
		MessageType: messages.CommitMessageType,
		PayLoad: &messages.CommitMessage{
			Commit: commit,
		},
	}
	msgRaw, err := messages.EncodeMessage(committed)
	assert.Nil(t, err)
	assert.NotNil(t, msgRaw)
	go fbft.Start()
	messages.Unicast(account, msgRaw, messages.CommitMessageType, mockHash)
	time.Sleep(1 * time.Second)
}

var fakeSignature = []byte{
	0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33,
	0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d,
}

func TestBftCore_receiveRequest(t *testing.T) {
	fbft := NewFBFTCore(mockAccounts[0], nil)
	fbft.timeoutTimer = time.NewTimer(30 * time.Second)
	assert.NotNil(t, fbft)
	fbft.nodes.peers = mockAccounts
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
	fbft.nodes.master = mockAccounts[1]
	fbft.receiveRequest(request)
	// absence of signature
	fbft.nodes.master = mockAccounts[0]
	fbft.receiveRequest(request)

	request.Payload.Header.SigData = append(request.Payload.Header.SigData, fakeSignature)

	var b *blockchain.BlockChain
	monkey.Patch(blockchain.NewBlockChainByBlockHash, func(types.Hash) (*blockchain.BlockChain, error) {
		return b, nil
	})
	var w *worker.Worker
	monkey.PatchInstanceMethod(reflect.TypeOf(w), "VerifyBlock", func(*worker.Worker) error {
		return nil
	})
	monkey.Patch(signature.Sign, func(signature.Signer, []byte) ([]byte, error) {
		return nil, fmt.Errorf("get signature failed")
	})
	fbft.receiveRequest(request)

	monkey.Patch(signature.Sign, func(signature.Signer, []byte) ([]byte, error) {
		return fakeSignature, nil
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
	fbft.receiveRequest(request)
	monkey.Unpatch(net.ResolveTCPAddr)
	monkey.Unpatch(net.DialTCP)
	monkey.Unpatch(signature.Sign)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(w), "VerifyBlock")
	monkey.UnpatchInstanceMethod(reflect.TypeOf(&c), "Write")
}

func TestNewFBFTCore_broadcast(t *testing.T) {
	fbft := NewFBFTCore(mockAccounts[0], nil)
	assert.NotNil(t, fbft)
	fbft.nodes.peers = mockAccounts
	// resolve error
	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, fmt.Errorf("resolve error")
	})
	messages.BroadcastPeers(nil, messages.ProposalMessageType, mockHash, fbft.nodes.peers)

	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, nil
	})
	monkey.Patch(net.DialTCP, func(string, *net.TCPAddr, *net.TCPAddr) (*net.TCPConn, error) {
		return nil, fmt.Errorf("dail error")
	})
	messages.BroadcastPeers(nil, messages.ProposalMessageType, mockHash, fbft.nodes.peers)

	var c net.TCPConn
	monkey.Patch(net.DialTCP, func(string, *net.TCPAddr, *net.TCPAddr) (*net.TCPConn, error) {
		return &c, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(&c), "Write", func(*net.TCPConn, []byte) (int, error) {
		return 0, nil
	})
	messages.BroadcastPeers(nil, messages.ProposalMessageType, mockHash, fbft.nodes.peers)
	monkey.Unpatch(net.ResolveTCPAddr)
	monkey.Unpatch(net.DialTCP)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(&c), "Write")
}

func TestBftCore_unicast(t *testing.T) {
	fbft := NewFBFTCore(mockAccounts[0], nil)
	assert.NotNil(t, fbft)
	fbft.nodes.peers = mockAccounts
	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, fmt.Errorf("resolve error")
	})
	err := messages.Unicast(fbft.nodes.peers[1], nil, messages.ProposalMessageType, mockHash)
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
	err = messages.Unicast(fbft.nodes.peers[1], nil, messages.ProposalMessageType, mockHash)
	assert.NotNil(t, err)
	monkey.Unpatch(net.ResolveTCPAddr)
	monkey.Unpatch(net.DialTCP)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(&c), "Write")
}

func TestBftCore_receiveProposal(t *testing.T) {
	fbft := NewFBFTCore(mockAccounts[0], nil)
	fbft.timeoutTimer = time.NewTimer(30 * time.Second)
	assert.NotNil(t, fbft)
	fbft.nodes.peers = mockAccounts
	fbft.nodes.master = mockAccounts[0]
	// master receive proposal
	proposal := &messages.Proposal{
		Timestamp: 1535414400,
		Payload: &types.Block{
			Header: &types.Header{
				Height:    0,
				MixDigest: mockHash,
			},
		},
	}
	fbft.receiveProposal(proposal)

	// verify failed: Get NewBlockChainByBlockHash failed
	fbft.nodes.local.Extension.Id = mockAccounts[0].Extension.Id + 1
	monkey.Patch(signature.Verify, func(keypair.PublicKey, []byte) (types.Address, error) {
		return mockAccounts[1].Address, nil
	})
	fbft.receiveProposal(proposal)

	monkey.Patch(signature.Verify, func(keypair.PublicKey, []byte) (types.Address, error) {
		return mockAccounts[0].Address, nil
	})
	var b *blockchain.BlockChain
	monkey.Patch(blockchain.NewBlockChainByBlockHash, func(types.Hash) (*blockchain.BlockChain, error) {
		return b, nil
	})
	var w *worker.Worker
	monkey.PatchInstanceMethod(reflect.TypeOf(w), "VerifyBlock", func(*worker.Worker) error {
		return fmt.Errorf("verify block failed")
	})
	fbft.receiveProposal(proposal)

	monkey.PatchInstanceMethod(reflect.TypeOf(w), "VerifyBlock", func(*worker.Worker) error {
		return nil
	})
	monkey.Patch(signature.Sign, func(signature.Signer, []byte) ([]byte, error) {
		return nil, fmt.Errorf("get signature failed")
	})
	fbft.receiveProposal(proposal)

	monkey.Patch(signature.Sign, func(signature.Signer, []byte) ([]byte, error) {
		return fakeSignature, nil
	})
	monkey.Patch(json.Marshal, func(interface{}) ([]byte, error) {
		return nil, fmt.Errorf("marshal proposal msg failed")
	})
	fbft.receiveProposal(proposal)

	monkey.Unpatch(json.Marshal)
	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, fmt.Errorf("resolve error")
	})
	fbft.receiveProposal(proposal)
	monkey.Unpatch(net.ResolveTCPAddr)
	monkey.Unpatch(blockchain.NewBlockChainByBlockHash)
	monkey.Unpatch(signature.Sign)
	monkey.Unpatch(signature.Verify)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(w), "VerifyBlock")
}

func TestFbftCore_receiveResponse(t *testing.T) {
	fbft := NewFBFTCore(mockAccounts[0], nil)
	fbft.nodes.peers = mockAccounts
	fbft.nodes.master = mockAccounts[0]
	fbft.tolerance = 1
	event := NewEvent()
	event.Subscribe(types.EventMasterChange, func(v interface{}) {
		log.Error("receive view change event.")
		return
	})
	fbft.eventCenter = event
	response := &messages.Response{
		Account:   mockAccounts[1],
		Timestamp: time.Now().Unix(),
		Digest:    mockHash,
		Signature: mockSignset[1],
	}
	content := fbft.consensusPlugin.Add(mockHash, nil)
	assert.NotNil(t, content)
	ok := content.AddSignature(mockAccounts[0], mockSignset[0])
	assert.Equal(t, true, ok)
	ok = content.AddSignature(mockAccounts[1], mockSignset[1])
	assert.Equal(t, true, ok)
	go fbft.waitResponse(mockHash)
	monkey.Patch(signature.Verify, func(_ keypair.PublicKey, sign []byte) (types.Address, error) {
		var address types.Address
		if bytes.Equal(sign[:], mockSignset[0]) {
			address = mockAccounts[0].Address
		}
		if bytes.Equal(sign[:], mockSignset[1]) {
			address = mockAccounts[1].Address
		}
		if bytes.Equal(sign[:], mockSignset[2]) {
			address = mockAccounts[2].Address
		}
		if bytes.Equal(sign[:], mockSignset[3]) {
			address = mockAccounts[3].Address
		}
		return address, nil
	})
	fbft.receiveResponse(response)
	ch := <-fbft.result
	assert.Equal(t, 2, len(ch.Signatures))
	assert.NotNil(t, ch.Result)
	assert.Equal(t, fmt.Errorf("signature not satisfy"), ch.Result)

	response = &messages.Response{
		Account:   mockAccounts[2],
		Timestamp: time.Now().Unix(),
		Digest:    mockHash,
		Signature: mockSignset[2],
	}
	content.SetState(common.InConsensus)
	go fbft.waitResponse(mockHash)
	fbft.receiveResponse(response)
	ch = <-fbft.result
	assert.Equal(t, len(mockSignset[:3]), len(ch.Signatures))
	assert.Nil(t, ch.Result)
	assert.Equal(t, common.ToConsensus, content.State())

	response = &messages.Response{
		Account:   mockAccounts[3],
		Timestamp: time.Now().Unix(),
		Digest:    mockHash,
		Signature: mockSignset[3],
	}
	content.SetState(common.InConsensus)
	go fbft.waitResponse(mockHash)
	fbft.receiveResponse(response)
	ch = <-fbft.result
	assert.Equal(t, len(mockSignset[:3]), len(ch.Signatures))
	monkey.Unpatch(signature.Verify)
}

func TestFbftCore_SendCommit(t *testing.T) {
	blockSwitch := make(chan interface{})
	fbft := NewFBFTCore(mockAccounts[0], blockSwitch)
	assert.NotNil(t, fbft)
	fbft.nodes.peers = mockAccounts
	block := &types.Block{
		HeaderHash: mockHash,
		Header: &types.Header{
			MixDigest: mockHash,
			SigData:   mockSignset,
		},
	}
	mockCommit := &messages.Commit{
		Account:    mockAccounts[0],
		Timestamp:  time.Now().Unix(),
		Digest:     block.Header.MixDigest,
		Signatures: block.Header.SigData,
		BlockHash:  mockHash,
		Result:     true,
	}
	var c net.TCPConn
	monkey.Patch(net.DialTCP, func(string, *net.TCPAddr, *net.TCPAddr) (*net.TCPConn, error) {
		return &c, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(&c), "Write", func(*net.TCPConn, []byte) (int, error) {
		return 0, nil
	})
	var b *blockchain.BlockChain
	monkey.Patch(blockchain.NewBlockChainByBlockHash, func(types.Hash) (*blockchain.BlockChain, error) {
		return b, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "GetBlockByHash", func(*blockchain.BlockChain, types.Hash) (*types.Block, error) {
		return block, nil
	})
	go fbft.sendCommit(mockCommit, block)
	blocks := <-blockSwitch
	assert.NotNil(t, blocks)
	assert.Equal(t, blocks.(*types.Block).HeaderHash, block.HeaderHash)
	monkey.Unpatch(blockchain.NewBlockChainByBlockHash)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(b), "GetBlockByHash")

	commit := &messages.Commit{
		Account:    mockAccounts[0],
		Timestamp:  time.Now().Unix(),
		Digest:     mockHash,
		Signatures: mockSignset,
		BlockHash:  mockHash,
		Result:     true,
	}
	committed := messages.Message{
		MessageType: messages.CommitMessageType,
		PayLoad: &messages.CommitMessage{
			Commit: commit,
		},
	}
	msgRaw, err := messages.EncodeMessage(committed)
	assert.Nil(t, err)
	assert.NotNil(t, msgRaw)

	msg, err := messages.DecodeMessage(committed.MessageType, msgRaw[12:])
	assert.Nil(t, err)
	payload := msg.PayLoad
	result := payload.(*messages.CommitMessage).Commit
	assert.NotNil(t, result)
	assert.Equal(t, commit, result)
}

func TestFBFTPolicy_commit(t *testing.T) {
	mockAccount := account.Account{
		Address: types.Address{0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68,
			0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  0,
			Url: "127.0.0.1:8080",
		},
	}
	var receive = make(chan interface{})
	core := NewFBFTCore(mockAccount, receive)
	block := &types.Block{
		Header: &types.Header{
			ChainID:       1,
			PrevBlockHash: MockHash,
			StateRoot:     MockHash,
			TxRoot:        MockHash,
			ReceiptsRoot:  MockHash,
			Height:        1,
			Timestamp:     uint64(time.Now().Unix()),
			SigData:       mockSignset[:4],
		},
		Transactions: make([]*types.Transaction, 0),
	}
	core.nodes.peers = append(core.nodes.peers, mockAccount)
	var b *blockchain.BlockChain
	monkey.Patch(blockchain.NewBlockChainByBlockHash, func(types.Hash) (*blockchain.BlockChain, error) {
		return b, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "GetBlockByHash", func(*blockchain.BlockChain, types.Hash) (*types.Block, error) {
		return block, nil
	})
	go core.commit(block, true)
	received := <-receive
	assert.NotNil(t, received)
	assert.Equal(t, received, block)
}

func TestFbftCore_ProcessEvent(t *testing.T) {
	blockSwitch := make(chan interface{})
	fbft := NewFBFTCore(mockAccounts[0], blockSwitch)
	fbft.timeoutTimer = time.NewTimer(30 * time.Second)
	fbft.nodes.peers = mockAccounts
	fbft.tolerance = 1
	event := NewEvent()
	event.Subscribe(types.EventMasterChange, func(v interface{}) {
		log.Error("receive view change event.")
		return
	})
	fbft.eventCenter = event
	mockViewNum := uint64(0)
	nodes := []account.Account{mockAccounts[1]}
	viewChangeReq := &messages.ViewChangeReq{
		Account:   mockAccounts[1],
		Nodes:     nodes,
		Timestamp: time.Now().Unix(),
		ViewNum:   mockViewNum,
	}
	fbft.ProcessEvent(viewChangeReq)

	monkey.Patch(messages.BroadcastPeersFilter, func([]byte, messages.MessageType, types.Hash, []account.Account, account.Account) {
		return
	})
	mockViewNum = uint64(1)
	viewChangeReq = &messages.ViewChangeReq{
		Account:   mockAccounts[1],
		Nodes:     []account.Account{mockAccounts[1]},
		Timestamp: time.Now().Unix(),
		ViewNum:   mockViewNum,
	}
	fbft.ProcessEvent(viewChangeReq)
	currentViewNum := fbft.viewChange.GetCurrentViewNum()
	assert.Equal(t, common.DefaultViewNum, currentViewNum)
	request := fbft.viewChange.GetRequestByViewNum(mockViewNum)
	receivedNodes := request.GetReceivedAccounts()
	assert.Equal(t, 2, len(receivedNodes))
	assert.Equal(t, common.Viewing, request.GetViewRequestState())

	viewChangeReq = &messages.ViewChangeReq{
		Account:   mockAccounts[2],
		Nodes:     []account.Account{mockAccounts[2]},
		Timestamp: time.Now().Unix(),
		ViewNum:   mockViewNum,
	}
	fbft.ProcessEvent(viewChangeReq)
	receivedNodes = request.GetReceivedAccounts()
	assert.Equal(t, 3, len(receivedNodes))
	assert.Equal(t, common.ViewEnd, request.GetViewRequestState())
	currentViewNum = fbft.viewChange.GetCurrentViewNum()
	assert.Equal(t, mockViewNum, currentViewNum)
}
