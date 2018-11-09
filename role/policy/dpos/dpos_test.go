package dpos

import (
	"fmt"
	"github.com/DSiSc/blockchain"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/participates/config"
	"github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/validator/tools/account"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
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
}

func TestNewDPOSPolicy(t *testing.T) {
	policy, err := NewDPOSPolicy(mockAccounts)
	assert.Nil(t, err)
	assert.NotNil(t, policy)
	assert.Equal(t, common.DPOS_POLICY, policy.name)
}

func TestDPOSPolicy_PolicyName(t *testing.T) {
	policy, _ := NewDPOSPolicy(mockAccounts)
	assert.Equal(t, common.DPOS_POLICY, policy.name)
	assert.Equal(t, policy.name, policy.PolicyName())
}

var participateConf = config.ParticipateConfig{
	PolicyName: common.DPOS_POLICY,
	Delegates:  4,
}

func TestDPOSPolicy_RoleAssignments(t *testing.T) {
	participate, err := participates.NewParticipates(participateConf)
	assert.NotNil(t, participate)
	assert.Nil(t, err)

	dposPolicy, err1 := NewDPOSPolicy(mockAccounts)
	assert.NotNil(t, dposPolicy)
	assert.Nil(t, err1)

	assignment, err2 := dposPolicy.RoleAssignments()
	assert.Nil(t, assignment)
	assert.Equal(t, fmt.Errorf("get NewLatestStateBlockChain failed"), err2)

	monkey.Patch(blockchain.NewLatestStateBlockChain, func() (*blockchain.BlockChain, error) {
		var temp blockchain.BlockChain
		return &temp, nil
	})
	var b *blockchain.BlockChain
	var height = uint64(0)
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "GetCurrentBlock", func(*blockchain.BlockChain) *types.Block {
		return &types.Block{
			Header: &types.Header{
				Height: height,
			},
		}
	})
	assignment, err3 := dposPolicy.RoleAssignments()
	assert.NotNil(t, assignment)
	assert.Nil(t, err3)
	address := dposPolicy.participates[height+1]
	assert.Equal(t, common.Master, dposPolicy.assignments[address])
}

func TestDPOSPolicy_AppointRole(t *testing.T) {
	participate, err := participates.NewParticipates(participateConf)
	assert.NotNil(t, participate)
	assert.Nil(t, err)

	dposPolicy, err1 := NewDPOSPolicy(mockAccounts)
	assert.NotNil(t, dposPolicy)
	assert.Nil(t, err1)

	monkey.Patch(blockchain.NewLatestStateBlockChain, func() (*blockchain.BlockChain, error) {
		var temp blockchain.BlockChain
		return &temp, nil
	})
	var b *blockchain.BlockChain
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "GetCurrentBlock", func(*blockchain.BlockChain) *types.Block {
		return &types.Block{
			Header: &types.Header{
				Height: 0,
			},
		}
	})
	assignment, err3 := dposPolicy.RoleAssignments()
	assert.NotNil(t, assignment)
	assert.Nil(t, err3)

	address := dposPolicy.participates[0]
	assert.Equal(t, common.Slave, dposPolicy.assignments[address])
	dposPolicy.AppointRole(0)
	assert.Equal(t, common.Master, dposPolicy.assignments[address])
}

func TestDPOSPolicy_GetRoles(t *testing.T) {
	participate, err := participates.NewParticipates(participateConf)
	assert.NotNil(t, participate)
	assert.Nil(t, err)

	dposPolicy, err1 := NewDPOSPolicy(mockAccounts)
	assert.NotNil(t, dposPolicy)
	assert.Nil(t, err1)

	monkey.Patch(blockchain.NewLatestStateBlockChain, func() (*blockchain.BlockChain, error) {
		var temp blockchain.BlockChain
		return &temp, nil
	})
	var b *blockchain.BlockChain
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "GetCurrentBlock", func(*blockchain.BlockChain) *types.Block {
		return &types.Block{
			Header: &types.Header{
				Height: 0,
			},
		}
	})
	assignment, err3 := dposPolicy.RoleAssignments()
	assert.NotNil(t, assignment)
	assert.Nil(t, err3)

	account0 := dposPolicy.participates[0]
	assert.Equal(t, common.Slave, dposPolicy.GetRoles(account0))

	account1 := dposPolicy.participates[1]
	assert.Equal(t, common.Master, dposPolicy.GetRoles(account1))

	var fakeAccount account.Account
	assert.Equal(t, common.UnKnown, dposPolicy.GetRoles(fakeAccount))
}
