package solo

import (
	"fmt"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/participates/config"
	"github.com/DSiSc/galaxy/participates/policy/solo"
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

func Test_NewSoloPolicy(t *testing.T) {
	asserts := assert.New(t)
	policy, err := NewSoloPolicy(nil)
	asserts.Nil(err)
	asserts.NotNil(policy)
	asserts.Equal(common.SOLO_POLICY, policy.name)
	asserts.Nil(policy.participates)
}

func mock_conf(policy string) config.ParticipateConfig {
	return config.ParticipateConfig{
		PolicyName: policy,
		Delegates:  4,
	}
}

func Test_RoleAssignments(t *testing.T) {
	asserts := assert.New(t)
	conf := mock_conf("solo")
	participate, err := participates.NewParticipates(conf)
	asserts.NotNil(participate)
	asserts.Nil(err)

	policy, err := NewSoloPolicy(participate)
	asserts.Nil(err)
	asserts.NotNil(policy)
	fmt.Println()

	var s *solo.SoloPolicy
	monkey.PatchInstanceMethod(reflect.TypeOf(s), "GetParticipates", func(*solo.SoloPolicy) ([]account.Account, error) {
		return nil, fmt.Errorf("get particioates failed")
	})
	roles, errs := policy.RoleAssignments()
	asserts.Nil(roles)
	asserts.Equal(fmt.Errorf("get participates failed"), errs)

	monkey.PatchInstanceMethod(reflect.TypeOf(s), "GetParticipates", func(*solo.SoloPolicy) ([]account.Account, error) {
		return mockAccounts, nil
	})
	roles, errs = policy.RoleAssignments()
	asserts.Nil(roles)
	asserts.Equal(fmt.Errorf("more than one participate"), errs)

	monkey.PatchInstanceMethod(reflect.TypeOf(s), "GetParticipates", func(*solo.SoloPolicy) ([]account.Account, error) {
		return mockAccounts[:1], nil
	})
	roles, errs = policy.RoleAssignments()
	asserts.Nil(errs)
	asserts.NotNil(roles)
	asserts.Equal(1, len(roles))
	asserts.Equal(common.Master, roles[mockAccounts[0]])
}

func TestSoloPolicy_GetRoles(t *testing.T) {
	asserts := assert.New(t)
	conf := mock_conf("solo")
	participate, err := participates.NewParticipates(conf)
	asserts.NotNil(participate)
	asserts.Nil(err)

	policy, err := NewSoloPolicy(participate)
	asserts.Nil(err)
	asserts.NotNil(policy)
	fmt.Println()
	roles, err := policy.RoleAssignments()
	asserts.Nil(err)
	asserts.NotNil(roles)

	accounts, err := participate.GetParticipates()
	asserts.Nil(err)
	asserts.Equal(1, len(accounts))
	role, err := policy.GetRoles(accounts[0])
	asserts.Nil(err)
	asserts.Equal(common.Master, role)

	policy.assignments = make(map[account.Account]common.Roler, 0)
	role, err = policy.GetRoles(accounts[0])
	asserts.Equal(common.AssignmentNotBeExecute, err)
	asserts.Equal(common.UnKnown, role)
}
