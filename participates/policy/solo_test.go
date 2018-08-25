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
	asserts := assert.New(t)
	policy, err := NewSoloPolicy()
	asserts.NotNil(policy)
	asserts.Nil(err)
	asserts.Equal(SOLO_POLICY, policy.name, "they should not be equal")
}

func Test_PolicyName(t *testing.T) {
	asserts := assert.New(t)
	policy := mock_NewSoloPolicy()
	policyName := policy.PolicyName()
	asserts.Equal(SOLO_POLICY, policyName, "they should not be equal")
}

func Test_GetParticipates(t *testing.T) {
	asserts := assert.New(t)
	policy := mock_NewSoloPolicy()
	address, err := policy.GetParticipates()
	asserts.NotNil(address)
	asserts.Nil(err)
	asserts.Equal(0, len(address), "they should not be equal")
}
