package bft

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	consensusConfig "github.com/DSiSc/galaxy/consensus/config"
	"github.com/DSiSc/galaxy/consensus/messages"
	tools "github.com/DSiSc/galaxy/consensus/utils"
	"github.com/DSiSc/galaxy/participates/config"
	commonr "github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/validator/tools/account"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func mock_conf(policy string) config.ParticipateConfig {
	return config.ParticipateConfig{
		PolicyName: policy,
		Delegates:  4,
	}
}

var timeout = consensusConfig.ConsensusTimeout{
	TimeoutToChangeView: int64(10000),
}

func TestNewBFTPolicy(t *testing.T) {
	bft, err := NewBFTPolicy(mockAccounts[0], timeout)
	assert.NotNil(t, bft)
	assert.Nil(t, err)
	assert.Equal(t, common.BftPolicy, bft.name)
	assert.NotNil(t, bft.bftCore)
	assert.Equal(t, mockAccounts[0].Extension.Id, bft.bftCore.local.Extension.Id)
}

func TestBFTPolicy_PolicyName(t *testing.T) {
	bft, _ := NewBFTPolicy(mockAccounts[0], timeout)
	assert.Equal(t, common.BftPolicy, bft.name)
	assert.Equal(t, bft.name, bft.PolicyName())
	assert.Equal(t, mockAccounts[0].Extension.Id, bft.bftCore.local.Extension.Id)
}

func mockRoleAssignment(master account.Account, accounts []account.Account) map[account.Account]commonr.Roler {
	delegates := len(accounts)
	assignments := make(map[account.Account]commonr.Roler, delegates)
	for _, delegate := range accounts {
		if delegate == master {
			assignments[delegate] = commonr.Master
		} else {
			assignments[delegate] = commonr.Slave
		}
	}
	return assignments
}

func TestBFTPolicy_Initialization(t *testing.T) {
	bft, err := NewBFTPolicy(mockAccounts[0], timeout)
	assert.NotNil(t, bft)
	assert.Nil(t, err)

	bft.Initialization(mockAccounts[3], mockAccounts, nil, true)
	assert.Equal(t, bft.bftCore.peers, mockAccounts)
	assert.Equal(t, bft.bftCore.tolerance, uint8((len(mockAccounts)-1)/3))
	assert.Equal(t, bft.bftCore.master, mockAccounts[3])
	assert.Equal(t, 0, len(bft.bftCore.validator))
	assert.Equal(t, 0, len(bft.bftCore.payloads))
	assert.NotNil(t, bft.bftCore.signature)
	assert.Equal(t, 0, len(bft.bftCore.signature.signMap))
	assert.Equal(t, 0, len(bft.bftCore.signature.signatures))
}

func TestBFTPolicy_Start(t *testing.T) {
	bft, _ := NewBFTPolicy(mockAccounts[0], timeout)
	var b *bftCore
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "Start", func(*bftCore, account.Account) {
		log.Info("pass it.")
		return
	})
	bft.Start()
	monkey.UnpatchInstanceMethod(reflect.TypeOf(b), "Start")
}

var mockConsensusResult = &messages.ConsensusResult{
	Signatures: mockSignset,
	Result:     nil,
}

func TestBFTPolicy_ToConsensus(t *testing.T) {
	bft, err := NewBFTPolicy(mockAccounts[0], timeout)
	assert.NotNil(t, bft)
	assert.Nil(t, err)
	bft.bftCore.peers = mockAccounts
	monkey.Patch(tools.SendEvent, func(tools.Receiver, tools.Event) {
		bft.result <- mockConsensusResult
	})
	var b *bftCore
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "SendCommit", func(*bftCore, *messages.Commit, *types.Block) {
		return
	})
	proposal := &common.Proposal{
		Block: &types.Block{
			Header: &types.Header{
				Height: 0,
			},
		},
	}
	assert.Equal(t, 0, len(proposal.Block.Header.SigData))
	err = bft.ToConsensus(proposal)
	assert.Nil(t, err)
	assert.Equal(t, len(mockSignset), len(proposal.Block.Header.SigData))
	assert.Equal(t, mockSignset, proposal.Block.Header.SigData)

	bft.timeout = time.Duration(2)
	monkey.Patch(tools.SendEvent, func(tools.Receiver, tools.Event) {
		return
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "SendCommit", func(*bftCore, *messages.Commit, *types.Block) {
		return
	})
	err = bft.ToConsensus(proposal)
	assert.Equal(t, fmt.Errorf("timeout for consensus"), err)
}

var MockHash = types.Hash{
	0x1d, 0xcf, 0x7, 0xba, 0xfc, 0x42, 0xb0, 0x8d, 0xfd, 0x23, 0x9c, 0x45, 0xa4, 0xb9, 0x38, 0xd,
	0x8d, 0xfe, 0x5d, 0x6f, 0xa7, 0xdb, 0xd5, 0x50, 0xc9, 0x25, 0xb1, 0xb3, 0x4, 0xdc, 0xc5, 0x1c,
}

func TestBFTPolicy_commit(t *testing.T) {
	mockAccount := account.Account{
		Address: types.Address{0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68,
			0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  0,
			Url: "127.0.0.1:8080",
		},
	}
	bft, err := NewBFTPolicy(mockAccount, timeout)
	go bft.Start()
	assert.NotNil(t, bft)
	assert.Nil(t, err)
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
	bft.bftCore.peers = append(bft.bftCore.peers, mockAccount)
	bft.commit(block, true)
}

func TestFBFTPolicy_GetConsensusResult(t *testing.T) {
	bft, err := NewBFTPolicy(mockAccounts[0], timeout)
	assert.Nil(t, err)
	bft.Initialization(mockAccounts[0], mockAccounts, nil, false)
	result := bft.GetConsensusResult()
	assert.Equal(t, uint64(0), result.View)
	assert.Equal(t, mockAccounts[0], result.Master)
	assert.Equal(t, len(mockAccounts), len(result.Participate))
}
