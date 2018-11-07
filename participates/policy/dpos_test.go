package policy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var delegates uint64 = 4

func TestDPOSPolicy_PolicyName(t *testing.T) {
	asserts := assert.New(t)
	dpos, err := NewDPOSPolicy(delegates)
	asserts.Nil(err)
	asserts.NotNil(dpos)
	asserts.Equal(DPOS_POLICY, dpos.PolicyName())
	asserts.Equal(delegates, dpos.members)
	asserts.Equal(0, len(dpos.participates))
}

func TestNewDPOSPolicy(t *testing.T) {
	asserts := assert.New(t)
	dpos, err := NewDPOSPolicy(delegates)
	asserts.Nil(err)
	asserts.NotNil(dpos)
	asserts.Equal(dpos.name, DPOS_POLICY)
}

func TestDPOSPolicy_GetParticipates(t *testing.T) {
	asserts := assert.New(t)
	dpos, err := NewDPOSPolicy(delegates)
	asserts.Nil(err)
	asserts.NotNil(dpos)
	participates, err := dpos.GetParticipates()
	asserts.Nil(err)
	asserts.Equal(delegates, dpos.members)
	asserts.Equal(dpos.members, uint64(len(dpos.participates)))
	asserts.Equal(participates, dpos.participates)
}
