package bft

import (
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewBFTPolicy(t *testing.T) {
	bft, err := NewBFTPolicy(nil)
	assert.NotNil(t, bft)
	assert.Nil(t, err)
	assert.Equal(t, common.BFT_POLICY, bft.name)
}

func TestBFTPolicy_PolicyName(t *testing.T) {
	bft, _ := NewBFTPolicy(nil)
	assert.Equal(t, common.BFT_POLICY, bft.name)
	assert.Equal(t, bft.name, bft.PolicyName())
}
