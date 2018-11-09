package solo

import (
	"github.com/DSiSc/craft/types"
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
	policy, err := NewSoloPolicy(mockAccounts)
	asserts.Nil(policy)
	asserts.NotNil(err)
	asserts.Errorf(err, "solo role policy only support one participate")

	policy, err = NewSoloPolicy(mockAccounts[:1])
	policyName := policy.PolicyName()
	asserts.Equal(common.SOLO_POLICY, policyName)
	asserts.Equal(policy.name, policyName)
}

func Test_RoleAssignments(t *testing.T) {
	asserts := assert.New(t)

	policy, err := NewSoloPolicy(mockAccounts[:1])
	asserts.Nil(err)
	asserts.NotNil(policy)

	roles, errs := policy.RoleAssignments()
	asserts.Nil(errs)
	asserts.NotNil(roles)
	asserts.Equal(1, len(roles))

	role := policy.GetRoles(mockAccounts[0])
	asserts.Equal(common.Master, role)

	role = policy.GetRoles(mockAccounts[1])
	asserts.Equal(common.UnKnown, role)

	policy.participates = mockAccounts
	roles, err = policy.RoleAssignments()
	asserts.Nil(roles)
	asserts.Errorf(err, "more than one participate")
}
