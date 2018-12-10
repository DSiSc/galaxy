package fbft

import (
	"fmt"
	"github.com/DSiSc/blockchain"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/messages"
	"github.com/DSiSc/galaxy/consensus/utils"
	"github.com/DSiSc/galaxy/participates/config"
	commonr "github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/validator/tools/account"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

var mockSignset = [][]byte{
	{0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
	{0x34, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
	{0x35, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
	{0x36, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
	{0x37, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
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

var sigChannel = make(chan *messages.ConsensusResult)

func mock_conf(policy string) config.ParticipateConfig {
	return config.ParticipateConfig{
		PolicyName: policy,
		Delegates:  4,
	}
}

var timeout = int64(10)

func TestNewfbftPolicy(t *testing.T) {
	fbft, err := NewFBFTPolicy(mockAccounts[0], timeout, nil)
	assert.NotNil(t, fbft)
	assert.Nil(t, err)
	assert.Equal(t, common.FBFT_POLICY, fbft.name)
	assert.NotNil(t, fbft.core)
	assert.Equal(t, mockAccounts[0].Extension.Id, fbft.core.nodes.local.Extension.Id)
}

func TestBFTPolicy_PolicyName(t *testing.T) {
	fbft, _ := NewFBFTPolicy(mockAccounts[0], timeout, nil)
	assert.Equal(t, common.FBFT_POLICY, fbft.name)
	assert.Equal(t, fbft.name, fbft.PolicyName())
	assert.Equal(t, mockAccounts[0].Extension.Id, fbft.core.nodes.local.Extension.Id)
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
	fbft, err := NewFBFTPolicy(mockAccounts[0], timeout, nil)
	assert.NotNil(t, fbft)
	assert.Nil(t, err)

	assignment := mockRoleAssignment(mockAccounts[3], mockAccounts)
	err = fbft.Initialization(assignment, mockAccounts, nil)
	assert.Equal(t, fbft.core.nodes.peers, mockAccounts)
	assert.Equal(t, fbft.core.tolerance, uint8((len(mockAccounts)-1)/3))
	assert.Equal(t, fbft.core.nodes.master, mockAccounts[3])

	assignment[mockAccounts[3]] = commonr.Slave
	err = fbft.Initialization(assignment, mockAccounts, nil)
	assert.Equal(t, err, fmt.Errorf("no master exist"))
}

func TestBFTPolicy_Start(t *testing.T) {
	fbft, _ := NewFBFTPolicy(mockAccounts[0], timeout, nil)
	var b *fbftCore
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "Start", func(*fbftCore) {
		log.Info("pass it.")
		return
	})
	fbft.Start()
	monkey.UnpatchInstanceMethod(reflect.TypeOf(b), "Start")
}

var mockConsensusResult = messages.ConsensusResult{
	Signatures: mockSignset,
	Result:     nil,
}

func TestBFTPolicy_ToConsensus(t *testing.T) {
	fbft, err := NewFBFTPolicy(mockAccounts[0], timeout, nil)
	assert.NotNil(t, fbft)
	assert.Nil(t, err)
	fbft.core.nodes.peers = mockAccounts
	var mockBlockSwitch = make(chan interface{})
	fbft.core.blockSwitch = mockBlockSwitch
	event := NewEvent()
	event.Subscribe(types.EventConsensusFailed, func(v interface{}) {
		log.Error("receive consensus failed event.")
		return
	})
	fbft.core.eventCenter = event
	monkey.Patch(utils.SendEvent, func(utils.Receiver, utils.Event) {
		fbft.core.result <- mockConsensusResult
	})
	monkey.Patch(messages.BroadcastPeersFilter, func([]byte, messages.MessageType, types.Hash, []account.Account, account.Account) {
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
	var b *blockchain.BlockChain
	monkey.Patch(blockchain.NewBlockChainByBlockHash, func(types.Hash) (*blockchain.BlockChain, error) {
		return b, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "GetBlockByHash", func(*blockchain.BlockChain, types.Hash) (*types.Block, error) {
		return proposal.Block, nil
	})
	go func() {
		err = fbft.ToConsensus(proposal)
		assert.Nil(t, err)
	}()
	block := <-mockBlockSwitch
	assert.NotNil(t, block)
	assert.Equal(t, len(mockSignset), len(proposal.Block.Header.SigData))
	assert.Equal(t, mockSignset, proposal.Block.Header.SigData)

	fbft.timeout = time.Duration(2)
	monkey.Patch(utils.SendEvent, func(utils.Receiver, utils.Event) {
		return
	})
	go func() {
		err = fbft.ToConsensus(proposal)
		assert.Equal(t, fmt.Errorf("timeout for consensus"), err)
	}()
	assert.Equal(t, len(mockSignset), len(proposal.Block.Header.SigData))
	assert.Equal(t, mockSignset, proposal.Block.Header.SigData)
}

var MockHash = types.Hash{
	0x1d, 0xcf, 0x7, 0xba, 0xfc, 0x42, 0xb0, 0x8d, 0xfd, 0x23, 0x9c, 0x45, 0xa4, 0xb9, 0x38, 0xd,
	0x8d, 0xfe, 0x5d, 0x6f, 0xa7, 0xdb, 0xd5, 0x50, 0xc9, 0x25, 0xb1, 0xb3, 0x4, 0xdc, 0xc5, 0x1c,
}

func TestFBFTPolicy_GetConsensusResult(t *testing.T) {
	var timeout = int64(10)
	fbft, err := NewFBFTPolicy(mockAccounts[0], timeout, nil)
	assert.Nil(t, err)

	fbft.core.nodes.master = mockAccounts[1]
	fbft.core.nodes.peers = mockAccounts
	result := fbft.GetConsensusResult()
	assert.Equal(t, uint64(0), result.View)
	assert.Equal(t, commonr.Master, result.Roles[mockAccounts[1]])
	assert.Equal(t, len(mockAccounts), len(result.Participate))
}
