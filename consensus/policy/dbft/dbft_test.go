package dbft

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	consensusConfig "github.com/DSiSc/galaxy/consensus/config"
	"github.com/DSiSc/galaxy/consensus/messages"
	"github.com/DSiSc/galaxy/consensus/utils"
	"github.com/DSiSc/galaxy/participates/config"
	roleCommon "github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/validator/tools/account"
	"github.com/stretchr/testify/assert"
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
			Url: "127.0.0.1:8080",
		},
	},
	account.Account{
		Address: types.Address{0x34, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68,
			0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  1,
			Url: "127.0.0.1:8081"},
	},
	account.Account{
		Address: types.Address{0x35, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  2,
			Url: "127.0.0.1:8082",
		},
	},

	account.Account{
		Address: types.Address{0x36, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  3,
			Url: "127.0.0.1:8083",
		},
	},
}

var mockHash = types.Hash{
	0xbd, 0x79, 0x1d, 0x4a, 0xf9, 0x64, 0x8f, 0xc3, 0x7f, 0x94, 0xeb, 0x36, 0x53, 0x19, 0xf6, 0xd0,
	0xa9, 0x78, 0x9f, 0x9c, 0x22, 0x47, 0x2c, 0xa7, 0xa6, 0x12, 0xa9, 0xca, 0x4, 0x13, 0xc1, 0x4,
}

var sigChannel = make(chan *messages.ConsensusResult)

func TestNewDBFTCore(t *testing.T) {
	dbft := NewDBFTCore(mockAccounts[0], sigChannel)
	assert.NotNil(t, dbft)
	assert.Equal(t, mockAccounts[0], dbft.local)
}

var mockSignset = [][]byte{
	{0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
	{0x34, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
	{0x35, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
	{0x36, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
	{0x37, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
}

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
	dbft, err := NewDBFTPolicy(mockAccounts[0], timeout)
	assert.NotNil(t, dbft)
	assert.Nil(t, err)
	assert.Equal(t, common.DbftPolicy, dbft.name)
	assert.NotNil(t, dbft.core)
	assert.Equal(t, mockAccounts[0].Extension.Id, dbft.core.local.Extension.Id)
}

func TestBFTPolicy_PolicyName(t *testing.T) {
	dbft, _ := NewDBFTPolicy(mockAccounts[0], timeout)
	assert.Equal(t, common.DbftPolicy, dbft.name)
	assert.Equal(t, dbft.name, dbft.PolicyName())
	assert.Equal(t, mockAccounts[0].Extension.Id, dbft.core.local.Extension.Id)
}

func mockRoleAssignment(master account.Account, accounts []account.Account) map[account.Account]roleCommon.Roler {
	delegates := len(accounts)
	assignments := make(map[account.Account]roleCommon.Roler, delegates)
	for _, delegate := range accounts {
		if delegate == master {
			assignments[delegate] = roleCommon.Master
		} else {
			assignments[delegate] = roleCommon.Slave
		}
	}
	return assignments
}

func TestBFTPolicy_Initialization(t *testing.T) {
	dbft, err := NewDBFTPolicy(mockAccounts[0], timeout)
	assert.NotNil(t, dbft)
	assert.Nil(t, err)
	dbft.Initialization(mockAccounts[3], mockAccounts, nil, true)
	assert.Equal(t, dbft.core.peers, mockAccounts)
	assert.Equal(t, dbft.core.tolerance, uint8((len(mockAccounts)-1)/3))
	assert.Equal(t, dbft.core.master, mockAccounts[3])
	assert.Equal(t, 0, len(dbft.core.validator))
	assert.Equal(t, 0, len(dbft.core.payloads))
	assert.NotNil(t, dbft.core.signature)
	assert.Equal(t, 0, len(dbft.core.signature.signMap))
	assert.Equal(t, 0, len(dbft.core.signature.signatures))
}

func TestBFTPolicy_Start(t *testing.T) {
	dbft, _ := NewDBFTPolicy(mockAccounts[0], timeout)
	var b *dbftCore
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "Start", func(*dbftCore, account.Account) {
		log.Info("pass it.")
		return
	})
	dbft.Start()
	monkey.UnpatchInstanceMethod(reflect.TypeOf(b), "Start")
}

var mockConsensusResult = &messages.ConsensusResult{
	Signatures: mockSignset,
	Result:     nil,
}

func TestBFTPolicy_ToConsensus(t *testing.T) {
	dbft, err := NewDBFTPolicy(mockAccounts[0], timeout)
	assert.NotNil(t, dbft)
	assert.Nil(t, err)
	dbft.core.peers = mockAccounts
	monkey.Patch(utils.SendEvent, func(utils.Receiver, utils.Event) {
		dbft.result <- mockConsensusResult
	})
	var b *dbftCore
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "SendCommit", func(*dbftCore, *messages.Commit, *types.Block) {
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
	err = dbft.ToConsensus(proposal)
	assert.Nil(t, err)
	assert.Equal(t, len(mockSignset), len(proposal.Block.Header.SigData))
	assert.Equal(t, mockSignset, proposal.Block.Header.SigData)

	dbft.timeout = time.Duration(2)
	monkey.Patch(utils.SendEvent, func(utils.Receiver, utils.Event) {
		return
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "SendCommit", func(*dbftCore, *messages.Commit, *types.Block) {
		return
	})
	err = dbft.ToConsensus(proposal)
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
	dbft, err := NewDBFTPolicy(mockAccount, timeout)
	go dbft.Start()
	assert.NotNil(t, dbft)
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
	dbft.core.peers = append(dbft.core.peers, mockAccount)
	dbft.commit(block, true)
}

func TestDBFTPolicy_GetConsensusResult(t *testing.T) {
	dbft, err := NewDBFTPolicy(mockAccounts[0], timeout)
	assert.Nil(t, err)

	dbft.Initialization(mockAccounts[0], mockAccounts, nil, false)
	dbft.core.views.viewNum = 1
	result := dbft.GetConsensusResult()
	assert.Equal(t, mockAccounts[0], result.Master)
	assert.Equal(t, len(mockAccounts), len(result.Participate))
	assert.Equal(t, uint64(1), result.View)
}
