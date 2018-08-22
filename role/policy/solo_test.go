package policy

import (
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/participates/config"
	"github.com/DSiSc/galaxy/role/common"
	txpoolc "github.com/DSiSc/txpool/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func mock_address(num int) []txpoolc.Address {
	to := make([]txpoolc.Address, num)
	for m := 0; m < num; m++ {
		for j := 0; j < txpoolc.AddressLength; j++ {
			to[m][j] = byte(m)
		}
	}
	return to
}

func mock_conf() *config.ParticipatePolicy {
	return &config.ParticipatePolicy{
		PolicyName: "solo",
	}
}

func Test_NewSoloPolicy(t *testing.T) {
	assert := assert.New(t)
	address := mock_address(1)[0]
	policy, err := NewSoloPolicy(nil, address)
	assert.Nil(err)
	assert.NotNil(policy)
	policyName := policy.PolicyName()
	assert.Equal(POLICY_NAME, policyName, "they should not be equal")
	assert.Equal(policy.name, policyName, "they should not be equal")
}

func Test_RoleAssignments(t *testing.T) {
	assert := assert.New(t)
	address := mock_address(1)[0]
	conf := mock_conf()
	p, err := participates.NewParticipatePolicy(conf)
	assert.Nil(err)
	assert.NotNil(p)

	policy, err := NewSoloPolicy(p, address)
	assert.Nil(err)
	assert.NotNil(policy)

	roles, errs := policy.RoleAssignments()
	assert.Nil(errs)
	assert.Nil(roles)

	roler := policy.GetRoles(address)
	assert.Equal(common.Master, roler)
}
