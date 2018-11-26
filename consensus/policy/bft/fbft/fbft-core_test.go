package fbft

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

func TestNewBFTCore(t *testing.T) {
	bft := NewFBFTCore(mockAccounts[0], sigChannel)
	assert.NotNil(t, bft)
	assert.Equal(t, mockAccounts[0], bft.local)
}

func TestBftCore_ProcessEvent(t *testing.T) {
	var sigChannel = make(chan *messages.ConsensusResult)
	fbft := NewFBFTCore(mockAccounts[0], sigChannel)
	assert.NotNil(t, fbft)
	id := mockAccounts[0].Extension.Id
	err := fbft.ProcessEvent(nil)
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

	fbft.peers = mockAccounts
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
	err = fbft.ProcessEvent(mock_request)
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
	fbft.master = id + 1
	err = fbft.ProcessEvent(mock_proposal)
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

	fbft.master = id
	mockResponse := &messages.Response{
		Account:   mockAccounts[0],
		Timestamp: time.Now().Unix(),
		Digest:    mockHash,
		Signature: mockSignset[0],
	}
	fbft.signature.AddSignature(fbft.peers[1], mockSignset[1])
	fbft.signature.AddSignature(fbft.peers[2], mockSignset[2])
	fbft.tolerance = uint8((len(fbft.peers) - 1) / 3)
	fbft.digest = mockHash
	go fbft.waitResponse()
	go func() {
		err = fbft.ProcessEvent(mockResponse)
		assert.Nil(t, err)
	}()
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
	monkey.UnpatchInstanceMethod(reflect.TypeOf(&c), "Write")
	monkey.Unpatch(blockchain.NewBlockChainByBlockHash)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(w), "VerifyBlock")
}

func TestBftCore_Start(t *testing.T) {
	fbft := NewFBFTCore(mockAccounts[0], sigChannel)
	assert.NotNil(t, fbft)
	var account = account.Account{
		Extension: account.AccountExtension{
			Url: "127.0.0.1:8080",
		},
	}
	var fakePayload = []byte{
		0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33,
		0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d,
	}
	go fbft.Start(account)
	messages.Unicast(account, fakePayload, "none", mockHash)
	time.Sleep(1 * time.Second)
}

var fakeSignature = []byte{
	0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33,
	0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d,
}

func TestBftCore_receiveRequest(t *testing.T) {
	fbft := NewFBFTCore(mockAccounts[0], sigChannel)
	id := mockAccounts[0].Extension.Id
	assert.NotNil(t, fbft)
	fbft.peers = mockAccounts
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
	fbft.master = id + 1
	fbft.receiveRequest(request)
	// absence of signature
	fbft.master = id
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
	//  marshal failed
	monkey.Patch(json.Marshal, func(interface{}) ([]byte, error) {
		return nil, fmt.Errorf("marshal proposal msg failed")
	})
	fbft.receiveRequest(request)
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
	fbft.receiveRequest(request)
	monkey.Unpatch(net.ResolveTCPAddr)
	monkey.Unpatch(net.DialTCP)
	monkey.Unpatch(signature.Sign)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(w), "VerifyBlock")
	monkey.UnpatchInstanceMethod(reflect.TypeOf(&c), "Write")
}

func TestNewFBFTCore_broadcast(t *testing.T) {
	fbft := NewFBFTCore(mockAccounts[0], sigChannel)
	assert.NotNil(t, fbft)
	fbft.peers = mockAccounts
	// resolve error
	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, fmt.Errorf("resolve error")
	})
	messages.BroadcastPeers(nil, messages.ProposalMessageType, mockHash, fbft.peers)

	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, nil
	})
	monkey.Patch(net.DialTCP, func(string, *net.TCPAddr, *net.TCPAddr) (*net.TCPConn, error) {
		return nil, fmt.Errorf("dail error")
	})
	messages.BroadcastPeers(nil, messages.ProposalMessageType, mockHash, fbft.peers)

	var c net.TCPConn
	monkey.Patch(net.DialTCP, func(string, *net.TCPAddr, *net.TCPAddr) (*net.TCPConn, error) {
		return &c, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(&c), "Write", func(*net.TCPConn, []byte) (int, error) {
		return 0, nil
	})
	messages.BroadcastPeers(nil, messages.ProposalMessageType, mockHash, fbft.peers)
	monkey.Unpatch(net.ResolveTCPAddr)
	monkey.Unpatch(net.DialTCP)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(&c), "Write")
}

func TestBftCore_unicast(t *testing.T) {
	fbft := NewFBFTCore(mockAccounts[0], sigChannel)
	assert.NotNil(t, fbft)
	fbft.peers = mockAccounts
	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, fmt.Errorf("resolve error")
	})
	err := messages.Unicast(fbft.peers[1], nil, messages.ProposalMessageType, mockHash)
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
	err = messages.Unicast(fbft.peers[1], nil, messages.ProposalMessageType, mockHash)
	assert.Nil(t, err)
	monkey.Unpatch(net.ResolveTCPAddr)
	monkey.Unpatch(net.DialTCP)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(&c), "Write")
}

func TestBftCore_receiveProposal(t *testing.T) {
	fbft := NewFBFTCore(mockAccounts[0], sigChannel)
	assert.NotNil(t, fbft)
	fbft.peers = mockAccounts
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
	fbft.local.Extension.Id = mockAccounts[0].Extension.Id + 1
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
	var bb types.Receipts
	r := common.NewReceipt(nil, false, uint64(10))
	bb = append(bb, r)
	monkey.PatchInstanceMethod(reflect.TypeOf(w), "GetReceipts", func(*worker.Worker) types.Receipts {
		return bb
	})
	monkey.Patch(signature.Sign, func(signature.Signer, []byte) ([]byte, error) {
		return nil, fmt.Errorf("get signature failed")
	})
	fbft.digest = proposal.Payload.Header.MixDigest
	fbft.receiveProposal(proposal)
	_, ok := fbft.validator[fbft.digest]
	assert.Equal(t, false, ok)

	monkey.Patch(signature.Sign, func(signature.Signer, []byte) ([]byte, error) {
		return fakeSignature, nil
	})
	monkey.Patch(json.Marshal, func(interface{}) ([]byte, error) {
		return nil, fmt.Errorf("marshal proposal msg failed")
	})
	fbft.receiveProposal(proposal)
	receipts := fbft.validator[fbft.digest].receipts
	assert.Equal(t, receipts, bb)

	monkey.Patch(json.Marshal, func(interface{}) ([]byte, error) {
		return nil, nil
	})
	monkey.Patch(net.ResolveTCPAddr, func(string, string) (*net.TCPAddr, error) {
		return nil, fmt.Errorf("resolve error")
	})
	fbft.receiveProposal(proposal)
	monkey.Unpatch(net.ResolveTCPAddr)
	monkey.Unpatch(json.Marshal)
	monkey.Unpatch(blockchain.NewBlockChainByBlockHash)
	monkey.Unpatch(signature.Sign)
	monkey.Unpatch(signature.Verify)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(w), "VerifyBlock")
}

func TestBftCore_receiveResponse(t *testing.T) {
	var sigChannel = make(chan *messages.ConsensusResult)
	fbft := NewFBFTCore(mockAccounts[0], sigChannel)
	assert.NotNil(t, fbft)
	fbft.peers = mockAccounts
	fbft.digest = mockHash
	response := &messages.Response{
		Account:   mockAccounts[1],
		Timestamp: time.Now().Unix(),
		Digest:    mockHash,
		Signature: mockSignset[2],
	}
	fbft.signature.AddSignature(mockAccounts[0], mockSignset[0])
	fbft.signature.AddSignature(mockAccounts[1], mockSignset[1])
	go fbft.waitResponse()
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

	response = &messages.Response{
		Account:   mockAccounts[2],
		Timestamp: time.Now().Unix(),
		Digest:    mockHash,
		Signature: mockSignset[2],
	}
	go fbft.waitResponse()
	fbft.commit = false
	fbft.receiveResponse(response)
	ch = <-fbft.result
	assert.Equal(t, len(mockSignset[:3]), len(ch.Signatures))

	response = &messages.Response{
		Account:   mockAccounts[3],
		Timestamp: time.Now().Unix(),
		Digest:    mockHash,
		Signature: mockSignset[3],
	}
	go fbft.waitResponse()
	fbft.commit = false
	fbft.receiveResponse(response)
	ch = <-fbft.result
	assert.Equal(t, len(mockSignset[:4]), len(ch.Signatures))
	monkey.Unpatch(signature.Verify)
}

func TestBftCore_ProcessEvent2(t *testing.T) {
	fbft := NewFBFTCore(mockAccounts[0], sigChannel)
	assert.NotNil(t, fbft)
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
	fbft.ProcessEvent(mockCommit)

	fbft.validator[mockHash] = &payloadSets{
		block: &types.Block{
			Header: &types.Header{
				Height:    2,
				MixDigest: mockHash,
			},
		},
	}
	fbft.ProcessEvent(mockCommit)

	fbft.validator[mockHash] = &payloadSets{
		block: &types.Block{
			Header: &types.Header{
				Height:    1,
				MixDigest: mockHash,
			},
		},
	}
	fbft.ProcessEvent(mockCommit)
	assert.Equal(t, 0, len(fbft.validator[mockHash].block.Header.SigData))

	var b *blockchain.BlockChain
	monkey.Patch(blockchain.NewBlockChainByBlockHash, func(types.Hash) (*blockchain.BlockChain, error) {
		return b, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "WriteBlockWithReceipts", func(*blockchain.BlockChain, *types.Block, []*types.Receipt) error {
		return fmt.Errorf("write failed")
	})
	fbft.ProcessEvent(mockCommit)
	assert.Equal(t, 0, len(fbft.validator[mockHash].block.Header.SigData))

	monkey.PatchInstanceMethod(reflect.TypeOf(b), "WriteBlockWithReceipts", func(*blockchain.BlockChain, *types.Block, []*types.Receipt) error {
		return nil
	})
	fbft.ProcessEvent(mockCommit)
	assert.Equal(t, len(mockSignset), len(fbft.validator[mockHash].block.Header.SigData))
	monkey.UnpatchAll()
}

func TestBftCore_SendCommit(t *testing.T) {
	fbft := NewFBFTCore(mockAccounts[0], sigChannel)
	assert.NotNil(t, fbft)
	fbft.peers = mockAccounts
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
	fbft.SendCommit(mockCommit, block)

	peers := fbft.getCommitOrder(nil, 0)
	successOrder := []account.Account{
		fbft.peers[0],
		fbft.peers[2],
		fbft.peers[3],
		fbft.peers[1],
	}
	assert.Equal(t, successOrder, peers)

	peers = fbft.getCommitOrder(fmt.Errorf("error"), 0)
	failedOrder := []account.Account{
		fbft.peers[1],
		fbft.peers[2],
		fbft.peers[3],
		fbft.peers[0],
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
