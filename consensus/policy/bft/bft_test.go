package bft

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/participates/config"
	commonr "github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/validator/tools/account"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func mock_conf(policy string) config.ParticipateConfig {
	return config.ParticipateConfig{
		PolicyName: policy,
		Delegates:  4,
	}
}

func TestNewBFTPolicy(t *testing.T) {
	bft, err := NewBFTPolicy(mockAccounts[0])
	assert.NotNil(t, bft)
	assert.Nil(t, err)
	assert.Equal(t, common.BFT_POLICY, bft.name)
	assert.NotNil(t, bft.bftCore)
	// assert.Equal(t, uint8((conf.Delegates-1)/3), bft.bftCore.tolerance)
	assert.Equal(t, mockAccounts[0].Extension.Id, bft.bftCore.local.Extension.Id)
	// assert.Equal(t, mockAccounts[1].Extension.Id, bft.bftCore.master)
}

func TestBFTPolicy_PolicyName(t *testing.T) {
	bft, _ := NewBFTPolicy(mockAccounts[0])
	assert.Equal(t, common.BFT_POLICY, bft.name)
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
	bft, err := NewBFTPolicy(mockAccounts[0])
	assert.NotNil(t, bft)
	assert.Nil(t, err)

	assignment := mockRoleAssignment(mockAccounts[3], mockAccounts)
	err = bft.Initialization(assignment, mockAccounts[:3])
	assert.Equal(t, err, fmt.Errorf("role and peers not in consistent"))

	assignment[mockAccounts[3]] = commonr.Master
	err = bft.Initialization(assignment, mockAccounts)
	assert.Equal(t, bft.bftCore.peers, mockAccounts)
	assert.Equal(t, bft.bftCore.tolerance, uint8((len(mockAccounts)-1)/3))
	assert.Equal(t, bft.bftCore.master, mockAccounts[3].Extension.Id)

	assignment[mockAccounts[3]] = commonr.Slave
	err = bft.Initialization(assignment, mockAccounts)
	assert.Equal(t, err, fmt.Errorf("no master"))
}

func TestBFTPolicy_Start(t *testing.T) {
	bft, _ := NewBFTPolicy(mockAccounts[0])
	var b *bftCore
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "Start", func(*bftCore, account.Account) {
		log.Info("pass it.")
		return
	})
	bft.Start()
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
	bft, err := NewBFTPolicy(local)
	assert.Nil(t, err)
	go bft.Start()
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
	bft.bftCore = NewBFTCore(uint64(0), ch)
	bft.bftCore.master = uint64(0)
	bft.bftCore.peers = []account.Account{local}
	bft.bftCore.tolerance = 0
	go tools.SendEvent(bft.bftCore, request)
	result := <-bft.result
	assert.NotNil(t, result)
}
*/
