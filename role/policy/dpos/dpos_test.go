package dpos

import (
	"fmt"
	"github.com/DSiSc/blockchain"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/participates/config"
	"github.com/DSiSc/galaxy/participates/policy/dpos"
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
	policy, err := NewDPOSPolicy(nil)
	assert.Nil(t, err)
	assert.NotNil(t, policy)
	assert.Equal(t, common.DPOS_POLICY, policy.name)
}

func TestDPOSPolicy_PolicyName(t *testing.T) {
	policy, _ := NewDPOSPolicy(nil)
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

	dposPolicy, err := NewDPOSPolicy(participate)
	assert.NotNil(t, dposPolicy)
	assert.Nil(t, err)

	var s *dpos.DPOSPolicy
	monkey.PatchInstanceMethod(reflect.TypeOf(s), "GetParticipates", func(*dpos.DPOSPolicy) ([]account.Account, error) {
		return nil, fmt.Errorf("get particioates failed")
	})
	assignment, err := dposPolicy.RoleAssignments()
	assert.Nil(t, assignment)
	assert.Equal(t, fmt.Errorf("get participates failed"), err)

	monkey.PatchInstanceMethod(reflect.TypeOf(s), "GetParticipates", func(*dpos.DPOSPolicy) ([]account.Account, error) {
		return mockAccounts, nil
	})
	assignment, err = dposPolicy.RoleAssignments()
	assert.Nil(t, assignment)
	assert.Equal(t, fmt.Errorf("get NewLatestStateBlockChain failed"), err)

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
	address := mockAccounts[height+1]
	assert.Equal(t, common.Master, dposPolicy.assignments[address])
}

func TestDPOSPolicy_AppointRole(t *testing.T) {
	participate, err := participates.NewParticipates(participateConf)
	assert.NotNil(t, participate)
	assert.Nil(t, err)

	dposPolicy, err1 := NewDPOSPolicy(participate)
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

	account0 := mockAccounts[0]
	assert.Equal(t, common.Slave, dposPolicy.assignments[account0])

	dposPolicy.AppointRole(account0)
	assert.Equal(t, common.Master, dposPolicy.assignments[account0])

	dposPolicy.assignments[account0] = common.Slave
	err = dposPolicy.AppointRole(account0)
	assert.Equal(t, fmt.Errorf("no master exist in current delegates"), err)

	var fakeAccount account.Account
	err = dposPolicy.AppointRole(fakeAccount)
	assert.Equal(t, fmt.Errorf("appoint account is not a delegate"), err)
}

func TestDPOSPolicy_GetRoles(t *testing.T) {
	participate, err := participates.NewParticipates(participateConf)
	assert.NotNil(t, participate)
	assert.Nil(t, err)

	dposPolicy, err := NewDPOSPolicy(participate)
	assert.NotNil(t, dposPolicy)
	assert.Nil(t, err)

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
	assignment, err := dposPolicy.RoleAssignments()
	assert.NotNil(t, assignment)
	assert.Nil(t, err)

	dposPolicy.assignments = make(map[account.Account]common.Roler)
	account0 := mockAccounts[0]
	role, err := dposPolicy.GetRoles(account0)
	assert.Equal(t, common.UnKnown, role)
	assert.Equal(t, common.AssignmentNotBeExecute, err)

	dposPolicy.assignments = assignment
	role, err = dposPolicy.GetRoles(account0)
	assert.Nil(t, err)
	assert.Equal(t, common.Slave, role)

	account1 := mockAccounts[1]
	role, err = dposPolicy.GetRoles(account1)
	assert.Equal(t, common.Master, role)

	var fakeAccount account.Account
	role, err = dposPolicy.GetRoles(fakeAccount)
	assert.Equal(t, common.UnKnown, role)
}
