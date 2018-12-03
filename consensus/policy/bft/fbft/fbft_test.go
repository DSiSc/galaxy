package fbft

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/policy/bft/messages"
	"github.com/DSiSc/galaxy/consensus/policy/bft/tools"
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
	fbft, err := NewFBFTPolicy(mockAccounts[0], timeout)
	assert.NotNil(t, fbft)
	assert.Nil(t, err)
	assert.Equal(t, common.FBFT_POLICY, fbft.name)
	assert.NotNil(t, fbft.core)
	assert.Equal(t, mockAccounts[0].Extension.Id, fbft.core.local.Extension.Id)
}

func TestBFTPolicy_PolicyName(t *testing.T) {
	fbft, _ := NewFBFTPolicy(mockAccounts[0], timeout)
	assert.Equal(t, common.FBFT_POLICY, fbft.name)
	assert.Equal(t, fbft.name, fbft.PolicyName())
	assert.Equal(t, mockAccounts[0].Extension.Id, fbft.core.local.Extension.Id)
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
	fbft, err := NewFBFTPolicy(mockAccounts[0], timeout)
	assert.NotNil(t, fbft)
	assert.Nil(t, err)

	assignment := mockRoleAssignment(mockAccounts[3], mockAccounts)
	err = fbft.Initialization(assignment, mockAccounts, nil)
	assert.Equal(t, fbft.core.peers, mockAccounts)
	assert.Equal(t, fbft.core.tolerance, uint8((len(mockAccounts)-1)/3))
	assert.Equal(t, fbft.core.master, mockAccounts[3].Extension.Id)
	assert.Equal(t, 0, len(fbft.core.validator))
	assert.NotNil(t, fbft.core.signature)
	assert.Equal(t, 0, len(fbft.core.signature.SignMap))
	assert.Equal(t, 0, len(fbft.core.signature.Signatures))

	assignment[mockAccounts[3]] = commonr.Slave
	err = fbft.Initialization(assignment, mockAccounts, nil)
	assert.Equal(t, err, fmt.Errorf("no master"))
}

func TestBFTPolicy_Start(t *testing.T) {
	fbft, _ := NewFBFTPolicy(mockAccounts[0], timeout)
	var b *fbftCore
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "Start", func(*fbftCore, account.Account) {
		log.Info("pass it.")
		return
	})
	fbft.Start()
	monkey.UnpatchInstanceMethod(reflect.TypeOf(b), "Start")
}

/*
func TestBFTPolicy_Halt(t *testing.T) {
	local := account.Account{
		Address: types.Address{0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  0,
			Url: "127.0.0.1:8080",
		},
	}
	fbft, err := NewBFTPolicy(local)
	assert.Nil(t, err)
	go fbft.Start()
	var sign = [][]byte{{0x33, 0x3c, 0x33, 0x10, 0x82}}
	time.Sleep(5*time.Second)
	request := &messages.Request{
		Timestamp: time.Now().Unix(),
		Payload:   &types.Block{
			Header:&types.Header{
				Height:1,
				SigData:sign,
			},
		},
	}
	var ch = make(chan messages.SignatureSet)
	fbft.core = NewBFTCore(uint64(0), ch)
	fbft.core.master = uint64(0)
	fbft.core.peers = []account.Account{local}
	fbft.core.tolerance = 0
	go tools.SendEvent(fbft.core, request)
	result := <-fbft.result
	assert.NotNil(t, result)
}
*/

var mockConsensusResult = &messages.ConsensusResult{
	Signatures: mockSignset,
	Result:     nil,
}

func TestBFTPolicy_ToConsensus(t *testing.T) {
	fbft, err := NewFBFTPolicy(mockAccounts[0], timeout)
	assert.NotNil(t, fbft)
	assert.Nil(t, err)
	fbft.core.peers = mockAccounts
	monkey.Patch(tools.SendEvent, func(tools.Receiver, tools.Event) {
		fbft.result <- mockConsensusResult
	})
	var b *fbftCore
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "SendCommit", func(*fbftCore, *messages.Commit, *types.Block) {
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
	err = fbft.ToConsensus(proposal)
	assert.Nil(t, err)
	assert.Equal(t, len(mockSignset), len(proposal.Block.Header.SigData))
	assert.Equal(t, mockSignset, proposal.Block.Header.SigData)

	fbft.timeout = time.Duration(2)
	monkey.Patch(tools.SendEvent, func(tools.Receiver, tools.Event) {
		return
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "SendCommit", func(*fbftCore, *messages.Commit, *types.Block) {
		return
	})
	err = fbft.ToConsensus(proposal)
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
	fbft, err := NewFBFTPolicy(mockAccount, timeout)
	go fbft.Start()
	assert.NotNil(t, fbft)
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
	fbft.core.peers = append(fbft.core.peers, mockAccount)
	fbft.commit(block, true)
}

func TestFBFTPolicy_GetConsensusResult(t *testing.T) {
	var timeout = int64(10)
	fbft, err := NewFBFTPolicy(mockAccounts[0], timeout)
	assert.Nil(t, err)

	fbft.core.master = 1
	fbft.core.peers = mockAccounts
	result := fbft.GetConsensusResult()
	assert.Equal(t, uint64(0), result.View)
	assert.Equal(t, commonr.Master, result.Roles[mockAccounts[1]])
	assert.Equal(t, len(mockAccounts), len(result.Participate))
}
