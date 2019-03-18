package solo

import (
	"github.com/DSiSc/galaxy/participates/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewSoloPolicy(t *testing.T) {
	asserts := assert.New(t)
	policy := NewSoloPolicy()
	asserts.NotNil(policy)
	asserts.Equal(common.SoloPolicy, policy.name, "they should not be equal")
}

func Test_PolicyName(t *testing.T) {
	asserts := assert.New(t)
	policy := NewSoloPolicy()
	policyName := policy.PolicyName()
	asserts.Equal(common.SoloPolicy, policyName, "they should not be equal")
}

func Test_GetParticipates(t *testing.T) {
	asserts := assert.New(t)
	policy := NewSoloPolicy()
	address, err := policy.GetParticipates()
	asserts.NotNil(address)
	asserts.Nil(err)
	asserts.Equal(1, len(address), "they should not be equal")
	asserts.Equal("127.0.0.1:8080", address[0].Extension.Url)
}
