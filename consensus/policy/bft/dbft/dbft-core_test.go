package dbft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/DSiSc/blockchain"
	"github.com/DSiSc/craft/types"
	commonc "github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/policy/bft/messages"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/validator/tools/account"
	"github.com/DSiSc/validator/tools/signature"
	"github.com/DSiSc/validator/tools/signature/keypair"
	"github.com/DSiSc/validator/worker"
	"github.com/DSiSc/validator/worker/common"
	"github.com/stretchr/testify/assert"
	"net"
	"reflect"
	"testing"
	"time"
)

func TestNewdbftCore(t *testing.T) {
	ddbft := NewDBFTCore(mockAccounts[0], sigChannel)
	assert.NotNil(t, ddbft)
	assert.Equal(t, mockAccounts[0], ddbft.local)
}

func TestDbftCore_ProcessEvent(t *testing.T) {
	var sigChannel = make(chan *messages.ConsensusResult)
	ddbft := NewDBFTCore(mockAccounts[0], sigChannel)
	assert.NotNil(t, ddbft)
	id := mockAccounts[0].Extension.Id
	err := ddbft.ProcessEvent(nil)
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

	ddbft.peers = mockAccounts
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
	err = ddbft.ProcessEvent(mock_request)
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
	ddbft.master = id + 1
	err = ddbft.ProcessEvent(mock_proposal)
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

	ddbft.master = id
	mockResponse := &messages.Response{
		Account:   mockAccounts[0],
		Timestamp: time.Now().Unix(),
		Digest:    mockHash,
		Signature: mockSignset[0],
	}
	ddbft.signature.addSignature(ddbft.peers[1], mockSignset[1])
	ddbft.signature.addSignature(ddbft.peers[2], mockSignset[2])
	ddbft.tolerance = uint8((len(ddbft.peers) - 1) / 3)
	ddbft.digest = mockHash
	go ddbft.waitResponse()
	go func() {
		err = ddbft.ProcessEvent(mockResponse)
		assert.Nil(t, err)
	}()
	ch := <-ddbft.result
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
	ddbft.ProcessEvent(mockCommit)
	monkey.Unpatch(net.ResolveTCPAddr)
	monkey.Unpatch(net.DialTCP)
	monkey.Unpatch(signature.Verify)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(&c), "Write")
	monkey.Unpatch(blockchain.NewBlockChainByBlockHash)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(w), "VerifyBlock")
}

func TestDbftCore_Start(t *testing.T) {
	ddbft := NewDBFTCore(mockAccounts[0], sigChannel)
	assert.NotNil(t, ddbft)
	var account = account.Account{
		Extension: account.AccountExtension{
			Url: "127.0.0.1:8080",
		},
	}
	var fakePayload = []byte{
		0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33,
		0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d,
	}
	go ddbft.Start(account)
	ddbft.unicast(account, fakePayload, "none", mockHash)
	time.Sleep(1 * time.Second)
}

var fakeSignature = []byte{
	0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33,
	0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d,
}

func TestDftCore_receiveRequest(t *testing.T) {
	dbft := NewDBFTCore(mockAccounts[0], sigChannel)
	id := mockAccounts[0].Extension.Id
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
	dbft.master = id + 1
	dbft.receiveRequest(request)
	// absence of signature
	dbft.master = id
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
	monkey.Unpatch(net.ResolveTCPAddr)
	monkey.Unpatch(net.DialTCP)
	monkey.Unpatch(signature.Sign)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(w), "VerifyBlock")
	monkey.UnpatchInstanceMethod(reflect.TypeOf(&c), "Write")
}

func TestNewDBFTCore_broadcast(t *testing.T) {
	dbft := NewDBFTCore(mockAccounts[0], sigChannel)
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
	monkey.Unpatch(net.ResolveTCPAddr)
	monkey.Unpatch(net.DialTCP)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(&c), "Write")
}

func TestDbftCore_unicast(t *testing.T) {
	dbft := NewDBFTCore(mockAccounts[0], sigChannel)
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
	monkey.Unpatch(net.ResolveTCPAddr)
	monkey.Unpatch(net.DialTCP)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(&c), "Write")
}

func TestDbftCore_receiveProposal(t *testing.T) {
	dbft := NewDBFTCore(mockAccounts[0], sigChannel)
	assert.NotNil(t, dbft)
	dbft.peers = mockAccounts

	// master receive proposal
	proposalHeight := uint64(2)
	proposal := &messages.Proposal{
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

	monkey.Patch(blockchain.NewLatestStateBlockChain, func() (*blockchain.BlockChain, error) {
		return b, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "GetCurrentBlockHeight", func(*blockchain.BlockChain) uint64 {
		return proposalHeight - 2
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
	r := common.NewReceipt(nil, false, uint64(10))
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
	assert.NotNil(t, dbft)
	dbft.peers = mockAccounts
	dbft.digest = mockHash
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
}

func TestDbftCore_ProcessEvent2(t *testing.T) {
	dbft := NewDBFTCore(mockAccounts[0], sigChannel)
	assert.NotNil(t, dbft)
	block0 := &types.Block{
		Header: &types.Header{
			Height:    1,
			MixDigest: mockHash,
			SigData:   mockSignset,
		},
	}
	hashBlock0 := commonc.HeaderHash(block0)
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
	committed := &messages.Message{
		MessageType: messages.CommitMessageType,
		Payload: &messages.CommitMessage{
			Commit: commit,
		},
	}
	msgRaw, err := json.Marshal(committed)
	assert.Nil(t, err)
	assert.NotNil(t, msgRaw)

	var msg messages.Message
	err = json.Unmarshal(msgRaw, &msg)
	payload := msg.Payload
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
	committed = &messages.Message{
		MessageType: messages.CommitMessageType,
		Payload: &messages.CommitMessage{
			Commit: commit,
		},
	}
	msgRaw, err = json.Marshal(committed)
	assert.Nil(t, err)
	assert.NotNil(t, msgRaw)

	err = json.Unmarshal(msgRaw, &msg)
	payload = msg.Payload
	result = payload.(*messages.CommitMessage).Commit
	assert.NotNil(t, result)
	assert.Equal(t, commit, result)
}
