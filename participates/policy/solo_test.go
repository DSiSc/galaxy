package policy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func mock_NewSoloPolicy() *SoloPolicy {
	policy, _ := NewSoloPolicy()
	return policy
}

func Test_NewSoloPolicy(t *testing.T) {
	assert := assert.New(t)
	policy, err := NewSoloPolicy()
	assert.NotNil(policy)
	assert.Nil(err)
	assert.Equal(SOLO_POLICY, policy.name, "they should not be equal")
}

func Test_PolicyName(t *testing.T) {
	assert := assert.New(t)
	policy := mock_NewSoloPolicy()
	policyName := policy.PolicyName()
	assert.Equal(SOLO_POLICY, policyName, "they should not be equal")
}

func Test_GetParticipates(t *testing.T) {
	assert := assert.New(t)
	policy := mock_NewSoloPolicy()
	address, err := policy.GetParticipates()
	assert.NotNil(address)
	assert.Nil(err)
	assert.Equal(0, len(address), "they should not be equal")
}
