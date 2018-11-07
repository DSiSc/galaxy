package policy

import (
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/participates/config"
	"github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/validator/tools/account"
	"github.com/stretchr/testify/assert"
	"testing"
)

func mock_address(num int) []types.Address {
	to := make([]types.Address, num)
	for m := 0; m < num; m++ {
		for j := 0; j < types.AddressLength; j++ {
			to[m][j] = byte(m)
		}
	}
	return to
}

func mock_conf() config.ParticipateConfig {
	return config.ParticipateConfig{
		PolicyName: "solo",
	}
}

var MockAccount = account.Account{
	Address: types.Address{
		0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33,
		0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d,
	},
}

func Test_NewSoloPolicy(t *testing.T) {
	asserts := assert.New(t)

	policy, err := NewSoloPolicy(nil, MockAccount)
	asserts.Nil(err)
	asserts.NotNil(policy)
	policyName := policy.PolicyName()
	asserts.Equal(common.SOLO_POLICY, policyName, "they should not be equal")
	asserts.Equal(policy.name, policyName, "they should not be equal")
	asserts.Nil(policy.participate)
	asserts.Equal(policy.local, MockAccount)
}

func Test_RoleAssignments(t *testing.T) {
	asserts := assert.New(t)

	var NotLocalAccount = account.Account{
		Address: types.Address{
			0x30, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33,
			0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d,
		},
	}

	conf := mock_conf()
	p, err := participates.NewParticipates(conf)
	asserts.Nil(err)
	asserts.NotNil(p)

	policy, err := NewSoloPolicy(p, MockAccount)
	asserts.Nil(err)
	asserts.NotNil(policy)

	roles, errs := policy.RoleAssignments()
	asserts.Nil(errs)
	asserts.NotNil(roles)

	roler := policy.GetRoles(MockAccount)
	asserts.Equal(common.Master, roler)

	roler = policy.GetRoles(NotLocalAccount)
	asserts.Equal(common.UnKnown, roler)
}
