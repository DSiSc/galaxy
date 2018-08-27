package policy

import (
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/participates/config"
	"github.com/DSiSc/galaxy/role/common"
	justitia_c "github.com/DSiSc/justitia/config"
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

func Test_NewSoloPolicy(t *testing.T) {
	asserts := assert.New(t)
	mock_node_name := "test"
	policy, err := NewSoloPolicy(nil, types.NodeAddress(mock_node_name))
	asserts.Nil(err)
	asserts.NotNil(policy)
	policyName := policy.PolicyName()
	asserts.Equal(SOLO_POLICY, policyName, "they should not be equal")
	asserts.Equal(policy.name, policyName, "they should not be equal")
	asserts.Nil(policy.participate)
	asserts.Equal(policy.local, types.NodeAddress("test"))
}

func Test_RoleAssignments(t *testing.T) {
	asserts := assert.New(t)
	address := types.NodeAddress(justitia_c.SINGLE_NODE_NAME)
	conf := mock_conf()
	p, err := participates.NewParticipates(conf)
	asserts.Nil(err)
	asserts.NotNil(p)

	policy, err := NewSoloPolicy(p, address)
	asserts.Nil(err)
	asserts.NotNil(policy)

	roles, errs := policy.RoleAssignments()
	asserts.Nil(errs)
	asserts.NotNil(roles)

	roler := policy.GetRoles(address)
	asserts.Equal(common.Master, roler)

	roler = policy.GetRoles(types.NodeAddress("test"))
	asserts.Equal(common.UnKnown, roler)
}
