package policy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDPOSPolicy_PolicyName(t *testing.T) {
	asserts := assert.New(t)
	dpos, err := NewDPOSPolicy()
	asserts.Nil(err)
	asserts.NotNil(dpos)
	asserts.Equal(dpos.PolicyName(), DPOS_POLICY)
}

func TestNewDPOSPolicy(t *testing.T) {
	asserts := assert.New(t)
	dpos, err := NewDPOSPolicy()
	asserts.Nil(err)
	asserts.NotNil(dpos)
	asserts.Equal(dpos.name, DPOS_POLICY)
}

func TestDPOSPolicy_ChangeParticipates(t *testing.T) {
	asserts := assert.New(t)
	dpos, err := NewDPOSPolicy()
	asserts.Nil(err)
	asserts.NotNil(dpos)
	err = dpos.ChangeParticipates()
	asserts.Nil(err)
}

func TestDPOSPolicy_GetParticipates(t *testing.T) {
	asserts := assert.New(t)
	dpos, err := NewDPOSPolicy()
	asserts.Nil(err)
	asserts.NotNil(dpos)
	participates, err := dpos.GetParticipates()
	asserts.Nil(err)
	asserts.NotEqual(0, len(participates))
}
