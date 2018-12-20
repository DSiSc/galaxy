package solo

import (
	"fmt"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/participates/config"
	"github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/validator/tools/account"
	"github.com/stretchr/testify/assert"
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

func Test_NewSoloPolicy(t *testing.T) {
	asserts := assert.New(t)
	policy, err := NewSoloPolicy()
	asserts.Nil(err)
	asserts.NotNil(policy)
	asserts.Equal(common.SoloPolicy, policy.name)
	asserts.Equal(0, len(policy.participates))
}

func mock_conf(policy string) config.ParticipateConfig {
	return config.ParticipateConfig{
		PolicyName: policy,
		Delegates:  4,
	}
}

func Test_RoleAssignments(t *testing.T) {
	asserts := assert.New(t)

	policy, err := NewSoloPolicy()
	asserts.Nil(err)
	asserts.NotNil(policy)

	roles, master, errs := policy.RoleAssignments(mockAccounts[:1])
	asserts.Nil(errs)
	asserts.NotNil(roles)
	asserts.Equal(1, len(roles))
	asserts.Equal(common.Master, roles[mockAccounts[0]])
	asserts.Equal(mockAccounts[0], master)

	policy.participates = mockAccounts
	roles, _, errs = policy.RoleAssignments(mockAccounts)
	asserts.Nil(roles)
	asserts.Equal(fmt.Errorf("more than one participate"), errs)
}

func TestSoloPolicy_GetRoles(t *testing.T) {
	asserts := assert.New(t)

	policy, err := NewSoloPolicy()
	asserts.Nil(err)
	asserts.NotNil(policy)
	fmt.Println()
	roles, master, err := policy.RoleAssignments(mockAccounts[:1])
	asserts.Nil(err)
	asserts.NotNil(roles)
	asserts.Equal(mockAccounts[0], master)

	role, err := policy.GetRoles(mockAccounts[0])
	asserts.Nil(err)
	asserts.Equal(common.Master, role)

	policy.assignments = make(map[account.Account]common.Roler, 0)
	role, err = policy.GetRoles(mockAccounts[0])
	asserts.Equal(common.AssignmentNotBeExecute, err)
	asserts.Equal(common.UnKnown, role)
}

func TestSoloPolicy_ChangeRoleAssignment(t *testing.T) {
	asserts := assert.New(t)

	policy, err := NewSoloPolicy()
	asserts.Nil(err)
	asserts.NotNil(policy)
	roles, master, err := policy.RoleAssignments(mockAccounts[:1])
	asserts.Equal(common.Master, roles[mockAccounts[0]])
	asserts.Equal(mockAccounts[0], master)

	asserts.Nil(err)
	asserts.NotNil(roles)
	policy.ChangeRoleAssignment(roles, uint64(1))
	asserts.Equal(common.Slave, roles[mockAccounts[0]])
}
