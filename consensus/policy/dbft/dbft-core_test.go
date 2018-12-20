package dbft

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
	"github.com/DSiSc/galaxy/consensus/utils"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/validator/tools/account"
	"github.com/DSiSc/validator/tools/signature"
	"github.com/DSiSc/validator/tools/signature/keypair"
	"github.com/DSiSc/validator/worker"
	vcommon "github.com/DSiSc/validator/worker/common"
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
	// TODO: open it when txswitch and blkswith stop complete
	//e.Subscribers = make(map[types.EventType]map[types.Subscriber]types.EventFunc)
	return
}

func TestNewdbftCore(t *testing.T) {
	dbft := NewDBFTCore(mockAccounts[0], sigChannel)
	assert.NotNil(t, dbft)
	assert.Equal(t, mockAccounts[0], dbft.local)
}

func TestDbftCore_ProcessEvent(t *testing.T) {
	id := 0
	var sigChannel = make(chan *messages.ConsensusResult)
	dbft := NewDBFTCore(mockAccounts[id], sigChannel)
	dbft.masterTimeout = time.NewTimer(10 * time.Second)
	assert.NotNil(t, dbft)
	err := dbft.ProcessEvent(nil)
	assert.Equal(t, fmt.Errorf("un support type <nil>"), err)

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

	dbft.peers = mockAccounts
	var mock_request = &messages.Request{
		Timestamp: time.Now().Unix(),
		Payload: &types.Block{
			Header: &types.Header{
				SigData: mockSignset[:1],
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
	err = dbft.ProcessEvent(mock_request)
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
	dbft.master = mockAccounts[id+1]
	err = dbft.ProcessEvent(mock_proposal)
	assert.Nil(t, err)

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

	dbft.master = mockAccounts[id]
	mockResponse := &messages.Response{
		Account:   mockAccounts[0],
		Timestamp: time.Now().Unix(),
		Digest:    mockHash,
		Signature: mockSignset[0],
	}
	dbft.signature.addSignature(dbft.peers[1], mockSignset[1])
	dbft.signature.addSignature(dbft.peers[2], mockSignset[2])
	dbft.tolerance = uint8((len(dbft.peers) - 1) / 3)
	dbft.digest = mockHash
	go dbft.waitResponse()
	go func() {
		err = dbft.ProcessEvent(mockResponse)
		assert.Nil(t, err)
	}()
	ch := <-dbft.result
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
	dbft.ProcessEvent(mockCommit)
	monkey.Unpatch(net.ResolveTCPAddr)
	monkey.Unpatch(net.DialTCP)
	monkey.Unpatch(signature.Verify)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(&c), "Write")
	monkey.Unpatch(blockchain.NewBlockChainByBlockHash)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(w), "VerifyBlock")
}

func TestDbftCore_Start(t *testing.T) {
	dbft := NewDBFTCore(mockAccounts[0], sigChannel)
	dbft.masterTimeout = time.NewTimer(10 * time.Second)
	assert.NotNil(t, dbft)
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
	go dbft.Start(account)
	messages.Unicast(account, msgRaw, messages.CommitMessageType, mockHash)
	time.Sleep(1 * time.Second)
}

var fakeSignature = []byte{
	0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33,
	0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d,
}

func TestDftCore_receiveRequest(t *testing.T) {
	id := 0
	dbft := NewDBFTCore(mockAccounts[id], sigChannel)
	dbft.masterTimeout = time.NewTimer(10 * time.Second)
	assert.NotNil(t, dbft)
	dbft.peers = mockAccounts
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
	dbft.master = mockAccounts[id+1]
	dbft.receiveRequest(request)
	// absence of signature
	dbft.master = mockAccounts[id]
	dbft.receiveRequest(request)

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
	dbft.receiveRequest(request)

	monkey.Patch(signature.Sign, func(signature.Signer, []byte) ([]byte, error) {
		return fakeSignature, nil
	})
	//  marshal failed
	monkey.Patch(json.Marshal, func(interface{}) ([]byte, error) {
		return nil, fmt.Errorf("marshal proposal msg failed")
	})
	dbft.receiveRequest(request)
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
	dbft.receiveRequest(request)
	monkey.UnpatchAll()
}

func TestNewDBFTCore_broadcast(t *testing.T) {
	dbft := NewDBFTCore(mockAccounts[0], sigChannel)
	dbft.masterTimeout = time.NewTimer(10 * time.Second)
	assert.NotNil(t, dbft)
	dbft.peers = mockAccounts
	// resolve error
	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, fmt.Errorf("resolve error")
	})
	dbft.broadcast(nil, messages.ProposalMessageType, mockHash)

	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, nil
	})
	monkey.Patch(net.DialTCP, func(string, *net.TCPAddr, *net.TCPAddr) (*net.TCPConn, error) {
		return nil, fmt.Errorf("dail error")
	})
	dbft.broadcast(nil, messages.ProposalMessageType, mockHash)

	var c net.TCPConn
	monkey.Patch(net.DialTCP, func(string, *net.TCPAddr, *net.TCPAddr) (*net.TCPConn, error) {
		return &c, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(&c), "Write", func(*net.TCPConn, []byte) (int, error) {
		return 0, nil
	})
	dbft.broadcast(nil, messages.ProposalMessageType, mockHash)
	monkey.UnpatchAll()
}

func TestDbftCore_unicast(t *testing.T) {
	dbft := NewDBFTCore(mockAccounts[0], sigChannel)
	dbft.masterTimeout = time.NewTimer(10 * time.Second)
	assert.NotNil(t, dbft)
	dbft.peers = mockAccounts
	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, fmt.Errorf("resolve error")
	})
	err := dbft.unicast(dbft.peers[1], nil, messages.ProposalMessageType, mockHash)
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
	err = dbft.unicast(dbft.peers[1], nil, messages.ProposalMessageType, mockHash)
	assert.Nil(t, err)
	monkey.UnpatchAll()
}

func TestDbftCore_receiveProposal(t *testing.T) {
	dbft := NewDBFTCore(mockAccounts[0], sigChannel)
	dbft.masterTimeout = time.NewTimer(10 * time.Second)
	assert.NotNil(t, dbft)
	dbft.peers = mockAccounts
	dbft.master = mockAccounts[0]

	// master receive proposal
	proposalHeight := uint64(2)
	proposal := &messages.Proposal{
		Account:   mockAccounts[0],
		Timestamp: 1535414400,
		Payload: &types.Block{
			Header: &types.Header{
				Height:    proposalHeight,
				MixDigest: mockHash,
			},
		},
	}
	dbft.receiveProposal(proposal)

	// verify failed
	dbft.local.Extension.Id = mockAccounts[0].Extension.Id + 1
	monkey.Patch(signature.Verify, func(keypair.PublicKey, []byte) (types.Address, error) {
		return mockAccounts[1].Address, nil
	})
	dbft.receiveProposal(proposal)

	monkey.Patch(signature.Verify, func(keypair.PublicKey, []byte) (types.Address, error) {
		return mockAccounts[0].Address, nil
	})
	var b *blockchain.BlockChain
	monkey.Patch(blockchain.NewLatestStateBlockChain, func() (*blockchain.BlockChain, error) {
		return b, fmt.Errorf("error")
	})
	dbft.receiveProposal(proposal)

	// current height less than proposal
	monkey.Patch(blockchain.NewLatestStateBlockChain, func() (*blockchain.BlockChain, error) {
		return b, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "GetCurrentBlockHeight", func(*blockchain.BlockChain) uint64 {
		return proposalHeight - 2
	})
	monkey.Patch(messages.Unicast, func(account.Account, []byte, messages.MessageType, types.Hash) error {
		return nil
	})
	dbft.receiveProposal(proposal)

	monkey.PatchInstanceMethod(reflect.TypeOf(b), "GetCurrentBlockHeight", func(*blockchain.BlockChain) uint64 {
		return proposalHeight
	})
	dbft.receiveProposal(proposal)

	monkey.PatchInstanceMethod(reflect.TypeOf(b), "GetCurrentBlockHeight", func(*blockchain.BlockChain) uint64 {
		return proposalHeight - 1
	})
	monkey.Patch(blockchain.NewBlockChainByBlockHash, func(types.Hash) (*blockchain.BlockChain, error) {
		return b, nil
	})
	var w *worker.Worker
	monkey.PatchInstanceMethod(reflect.TypeOf(w), "VerifyBlock", func(*worker.Worker) error {
		return fmt.Errorf("verify block failed")
	})
	dbft.receiveProposal(proposal)

	monkey.PatchInstanceMethod(reflect.TypeOf(w), "VerifyBlock", func(*worker.Worker) error {
		return nil
	})
	var bb types.Receipts
	r := vcommon.NewReceipt(nil, false, uint64(10))
	bb = append(bb, r)
	monkey.PatchInstanceMethod(reflect.TypeOf(w), "GetReceipts", func(*worker.Worker) types.Receipts {
		return bb
	})
	monkey.Patch(signature.Sign, func(signature.Signer, []byte) ([]byte, error) {
		return nil, fmt.Errorf("get signature failed")
	})
	dbft.digest = proposal.Payload.Header.MixDigest
	dbft.receiveProposal(proposal)
	_, ok := dbft.validator[dbft.digest]
	assert.Equal(t, false, ok)

	monkey.Patch(signature.Sign, func(signature.Signer, []byte) ([]byte, error) {
		return fakeSignature, nil
	})
	monkey.Patch(json.Marshal, func(interface{}) ([]byte, error) {
		return nil, fmt.Errorf("marshal proposal msg failed")
	})
	dbft.receiveProposal(proposal)
	receipts := dbft.validator[dbft.digest].receipts
	assert.Equal(t, receipts, bb)

	monkey.Patch(json.Marshal, func(interface{}) ([]byte, error) {
		return nil, nil
	})
	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, fmt.Errorf("resolve error")
	})
	dbft.receiveProposal(proposal)
	monkey.UnpatchAll()
}

func TestDbftCore_receiveResponse(t *testing.T) {
	var sigChannel = make(chan *messages.ConsensusResult)
	dbft := NewDBFTCore(mockAccounts[0], sigChannel)
	dbft.masterTimeout = time.NewTimer(10 * time.Second)
	assert.NotNil(t, dbft)
	dbft.peers = mockAccounts
	dbft.digest = mockHash
	dbft.master = mockAccounts[0]
	response := &messages.Response{
		Account:   mockAccounts[1],
		Timestamp: time.Now().Unix(),
		Digest:    mockHash,
		Signature: mockSignset[2],
	}
	dbft.signature.addSignature(mockAccounts[0], mockSignset[0])
	dbft.signature.addSignature(mockAccounts[1], mockSignset[1])
	go dbft.waitResponse()
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
	dbft.receiveResponse(response)
	ch := <-dbft.result
	assert.Equal(t, 2, len(ch.Signatures))

	response = &messages.Response{
		Account:   mockAccounts[2],
		Timestamp: time.Now().Unix(),
		Digest:    mockHash,
		Signature: mockSignset[2],
	}
	go dbft.waitResponse()
	dbft.commit = false
	dbft.receiveResponse(response)
	ch = <-dbft.result
	assert.Equal(t, len(mockSignset[:3]), len(ch.Signatures))

	response = &messages.Response{
		Account:   mockAccounts[3],
		Timestamp: time.Now().Unix(),
		Digest:    mockHash,
		Signature: mockSignset[3],
	}
	go dbft.waitResponse()
	dbft.commit = false
	dbft.receiveResponse(response)
	ch = <-dbft.result
	assert.Equal(t, len(mockSignset[:4]), len(ch.Signatures))
	monkey.Unpatch(signature.Verify)
	monkey.UnpatchAll()
}

func TestDbftCore_ProcessEvent2(t *testing.T) {
	dbft := NewDBFTCore(mockAccounts[0], sigChannel)
	dbft.masterTimeout = time.NewTimer(10 * time.Second)
	assert.NotNil(t, dbft)
	block0 := &types.Block{
		Header: &types.Header{
			Height:    1,
			MixDigest: mockHash,
			SigData:   mockSignset,
		},
	}
	hashBlock0 := common.HeaderHash(block0)
	mockCommit := &messages.Commit{
		Account:    mockAccounts[0],
		Timestamp:  time.Now().Unix(),
		Digest:     mockHash,
		Signatures: mockSignset,
		BlockHash:  hashBlock0,
		Result:     true,
	}
	dbft.ProcessEvent(mockCommit)

	dbft.validator[mockHash] = &payloadSets{
		block: &types.Block{
			Header: &types.Header{
				Height:    2,
				MixDigest: mockHash,
			},
		},
	}
	dbft.ProcessEvent(mockCommit)

	dbft.validator[mockHash] = &payloadSets{
		block: &types.Block{
			Header: &types.Header{
				Height:    1,
				MixDigest: mockHash,
			},
		},
	}
	dbft.ProcessEvent(mockCommit)
	assert.Equal(t, 0, len(dbft.validator[mockHash].block.Header.SigData))

	var b *blockchain.BlockChain
	monkey.Patch(blockchain.NewBlockChainByBlockHash, func(types.Hash) (*blockchain.BlockChain, error) {
		return b, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "WriteBlockWithReceipts", func(*blockchain.BlockChain, *types.Block, []*types.Receipt) error {
		return fmt.Errorf("write failed")
	})
	dbft.ProcessEvent(mockCommit)
	assert.Equal(t, 0, len(dbft.validator[mockHash].block.Header.SigData))

	monkey.PatchInstanceMethod(reflect.TypeOf(b), "WriteBlockWithReceipts", func(*blockchain.BlockChain, *types.Block, []*types.Receipt) error {
		return nil
	})
	dbft.ProcessEvent(mockCommit)
	assert.Equal(t, len(mockSignset), len(dbft.validator[mockHash].block.Header.SigData))
	monkey.UnpatchAll()
}

func TestDbftCore_SendCommit(t *testing.T) {
	dbft := NewDBFTCore(mockAccounts[0], sigChannel)
	dbft.masterTimeout = time.NewTimer(10 * time.Second)
	assert.NotNil(t, dbft)
	dbft.peers = mockAccounts
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
	dbft.SendCommit(mockCommit, block)

	peers := dbft.getCommitOrder(nil, 0)
	successOrder := []account.Account{
		dbft.peers[0],
		dbft.peers[2],
		dbft.peers[3],
		dbft.peers[1],
	}
	assert.Equal(t, successOrder, peers)

	peers = dbft.getCommitOrder(fmt.Errorf("error"), 0)
	failedOrder := []account.Account{
		dbft.peers[1],
		dbft.peers[2],
		dbft.peers[3],
		dbft.peers[0],
	}
	assert.Equal(t, failedOrder, peers)

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
	assert.NotNil(t, msg)
	payload := msg.PayLoad
	result := payload.(*messages.CommitMessage).Commit
	assert.NotNil(t, result)
	assert.Equal(t, commit, result)

	commit = &messages.Commit{
		Account:    mockAccounts[0],
		Timestamp:  time.Now().Unix(),
		Digest:     mockHash,
		Signatures: mockSignset,
		BlockHash:  mockHash,
		Result:     false,
	}
	committed = messages.Message{
		MessageType: messages.CommitMessageType,
		PayLoad: &messages.CommitMessage{
			Commit: commit,
		},
	}
	msgRaw, err = messages.EncodeMessage(committed)
	assert.Nil(t, err)
	assert.NotNil(t, msgRaw)

	msg, err = messages.DecodeMessage(committed.MessageType, msgRaw[12:])
	assert.Nil(t, err)
	assert.NotNil(t, msg)
	result = msg.PayLoad.(*messages.CommitMessage).Commit
	assert.NotNil(t, result)
	assert.Equal(t, commit, result)
	monkey.UnpatchAll()
}

func TestDbftCore_ProcessEvent3(t *testing.T) {
	masterAccount := mockAccounts[0]
	slaveAccount := mockAccounts[1]
	dbft := NewDBFTCore(mockAccounts[0], sigChannel)
	assert.NotNil(t, dbft)
	dbft.masterTimeout = time.NewTimer(10 * time.Second)
	dbft.peers = mockAccounts
	go dbft.Start(mockAccounts[0])
	var currentHeight uint64 = 1
	syncBlockMessage := messages.Message{
		MessageType: messages.SyncBlockReqMessageType,
		PayLoad: &messages.SyncBlockReqMessage{
			SyncBlockReq: &messages.SyncBlockReq{
				Account:    slaveAccount,
				Timestamp:  time.Now().Unix(),
				BlockStart: currentHeight + 1,
				BlockEnd:   currentHeight + 2,
			},
		},
	}
	msgRaw, err := messages.EncodeMessage(syncBlockMessage)
	assert.Nil(t, err)
	assert.NotNil(t, msgRaw)
	var b *blockchain.BlockChain
	monkey.Patch(blockchain.NewLatestStateBlockChain, func() (*blockchain.BlockChain, error) {
		return b, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "GetBlockByHeight",
		func(block *blockchain.BlockChain, height uint64) (*types.Block, error) {
			if currentHeight+1 == height {
				return &types.Block{
					Header: &types.Header{
						Height: currentHeight + 1,
					},
				}, nil
			}
			if currentHeight+2 == height {
				return &types.Block{
					Header: &types.Header{
						Height: currentHeight + 2,
					},
				}, nil
			}
			return nil, fmt.Errorf("not reach")
		})
	err = messages.Unicast(masterAccount, msgRaw, messages.SyncBlockReqMessageType, mockHash)
	time.Sleep(2 * time.Second)
	monkey.UnpatchAll()
}

func mockBlocks(blockNum int) []*types.Block {
	blocks := make([]*types.Block, 0)
	temp := &types.Block{
		Header: &types.Header{
			Height:    0,
			Timestamp: uint64(time.Now().Unix()),
		},
	}
	temp.HeaderHash = common.HeaderHash(temp)
	for index := 0; index < blockNum; index++ {
		block := &types.Block{
			Header: &types.Header{
				Height:        uint64(index + 1),
				PrevBlockHash: common.HeaderHash(temp),
			},
		}
		block.HeaderHash = common.HeaderHash(block)
		blocks = append(blocks, block)
	}
	return blocks[0:]
}

// test receiveSyncBlockResp
func TestDbftCore_ProcessEvent4(t *testing.T) {
	core := NewDBFTCore(mockAccounts[0], sigChannel)
	core.masterTimeout = time.NewTimer(10 * time.Second)
	syncBlockResp := &messages.SyncBlockResp{
		Blocks: mockBlocks(2),
	}
	var b *blockchain.BlockChain
	monkey.Patch(blockchain.NewBlockChainByBlockHash, func(types.Hash) (*blockchain.BlockChain, error) {
		return b, nil
	})
	var w *worker.Worker
	monkey.PatchInstanceMethod(reflect.TypeOf(w), "VerifyBlock", func(*worker.Worker) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "EventWriteBlockWithReceipts", func(*blockchain.BlockChain, *types.Block, []*types.Receipt, bool) error {
		return nil
	})
	utils.SendEvent(core, syncBlockResp)
	monkey.UnpatchAll()
}

// test receiveSyncBlockResp
func TestDbftCore_ProcessEvent5(t *testing.T) {
	core := NewDBFTCore(mockAccounts[0], sigChannel)
	core.masterTimeout = time.NewTimer(10 * time.Second)
	syncBlockReq := &messages.SyncBlockReq{
		Account:    mockAccounts[1],
		Timestamp:  time.Now().Unix(),
		BlockStart: uint64(1),
		BlockEnd:   uint64(2),
	}
	var b *blockchain.BlockChain
	monkey.Patch(blockchain.NewLatestStateBlockChain, func() (*blockchain.BlockChain, error) {
		return b, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "GetBlockByHeight", func(_ *blockchain.BlockChain, height uint64) (*types.Block, error) {
		return &types.Block{
			Header: &types.Header{
				Height: height,
			},
		}, nil
	})
	monkey.Patch(messages.Unicast, func(account.Account, []byte, messages.MessageType, types.Hash) error {
		return nil
	})
	utils.SendEvent(core, syncBlockReq)
	monkey.UnpatchAll()
}

func TestDbftCore_ProcessEvent6(t *testing.T) {
	core := NewDBFTCore(mockAccounts[0], sigChannel)
	core.masterTimeout = time.NewTimer(10 * time.Second)
	core.peers = mockAccounts
	core.tolerance = 1
	core.master = mockAccounts[2]
	currentHeight := uint64(2)
	event := NewEvent()
	event.Subscribe(types.EventMasterChange, func(v interface{}) {
		log.Error("receive view change event.")
		return
	})
	core.eventCenter = event
	viewChangeReq := &messages.ViewChangeReq{
		Nodes:     mockAccounts[:1],
		Timestamp: time.Now().Unix(),
		ViewNum:   currentHeight + 1,
	}
	var b *blockchain.BlockChain
	monkey.Patch(blockchain.NewLatestStateBlockChain, func() (*blockchain.BlockChain, error) {
		return b, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "GetCurrentBlockHeight", func(*blockchain.BlockChain) uint64 {
		return currentHeight
	})
	utils.SendEvent(core, viewChangeReq)

	monkey.Patch(messages.BroadcastPeers, func([]byte, messages.MessageType, types.Hash, []account.Account) {
		return
	})
	viewChangeReq = &messages.ViewChangeReq{
		Nodes:     mockAccounts[:2],
		Timestamp: time.Now().Unix(),
		ViewNum:   currentHeight + 2,
	}
	utils.SendEvent(core, viewChangeReq)

	viewChangeReq = &messages.ViewChangeReq{
		Nodes:     mockAccounts[:3],
		Timestamp: time.Now().Unix(),
		ViewNum:   currentHeight + 2,
	}
	utils.SendEvent(core, viewChangeReq)
	monkey.UnpatchAll()
}

func TestDbftCore_ProcessEvent7(t *testing.T) {
	core := NewDBFTCore(mockAccounts[0], sigChannel)
	core.masterTimeout = time.NewTimer(10 * time.Second)
	core.peers = mockAccounts
	core.tolerance = 1
	core.master = mockAccounts[3]
	event := NewEvent()
	event.Subscribe(types.EventMasterChange, func(v interface{}) {
		log.Error("receive view change event.")
		return
	})
	core.eventCenter = event

	// receive change view request from node 1
	assert.Equal(t, uint64(0), core.views.viewNum)
	viewChangeReq := &messages.ViewChangeReq{
		Account:   mockAccounts[1],
		Nodes:     mockAccounts[1:2],
		Timestamp: time.Now().Unix(),
		ViewNum:   uint64(1),
	}
	monkey.Patch(messages.BroadcastPeers, func([]byte, messages.MessageType, types.Hash, []account.Account) {
		return
	})
	core.receiveChangeViewReq(viewChangeReq)

	assert.Equal(t, uint64(0), core.views.viewNum)
	assert.Equal(t, common.ViewChanging, core.views.status)
	assert.Equal(t, 2, len(core.views.viewSets[viewChangeReq.ViewNum].requestNodes))
	assert.Equal(t, common.Viewing, core.views.viewSets[viewChangeReq.ViewNum].status)

	viewChangeReq = &messages.ViewChangeReq{
		Account:   mockAccounts[1],
		Nodes:     mockAccounts[0:1],
		Timestamp: time.Now().Unix(),
		ViewNum:   uint64(1),
	}
	core.receiveChangeViewReq(viewChangeReq)
	assert.Equal(t, uint64(0), core.views.viewNum)
	assert.Equal(t, common.ViewChanging, core.views.status)
	assert.Equal(t, 2, len(core.views.viewSets[viewChangeReq.ViewNum].requestNodes))
	assert.Equal(t, common.Viewing, core.views.viewSets[viewChangeReq.ViewNum].status)

	viewChangeReq = &messages.ViewChangeReq{
		Account:   mockAccounts[1],
		Nodes:     mockAccounts[0:3],
		Timestamp: time.Now().Unix(),
		ViewNum:   uint64(1),
	}
	core.receiveChangeViewReq(viewChangeReq)
	assert.Equal(t, uint64(1), core.views.viewNum)
	assert.Equal(t, common.ViewChanging, core.views.status)
	assert.Equal(t, 3, len(core.views.viewSets[viewChangeReq.ViewNum].requestNodes))
	assert.Equal(t, common.ViewEnd, core.views.viewSets[viewChangeReq.ViewNum].status)
	assert.Equal(t, mockAccounts[0], core.master)
}
