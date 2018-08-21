package policy

import (
	"github.com/DSiSc/galaxy/consensus/common"
	producer_c "github.com/DSiSc/producer/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewSoloPolicy(t *testing.T) {
	assert := assert.New(t)
	sp, err := NewSoloPolicy(nil)
	assert.Nil(err)
	assert.NotNil(sp)
	assert.Equal(POLICY_NAME, sp.name)
	assert.Nil(sp.participates)
	assert.Equal(POLICY_NAME, sp.PolicyName())
}

func mock_proposal() *common.Proposal {
	var block producer_c.Block
	return &common.Proposal{
		Block: &block,
	}
}

func mock_solo_proposal() *SoloProposal {
	return &SoloProposal{
		propoasl: nil,
		version:  0,
		status:   common.Proposing,
	}
}

func Test_toSoloProposal(t *testing.T) {
	assert := assert.New(t)
	p := mock_proposal()
	proposal := toSoloProposal(p)
	assert.NotNil(proposal)
	assert.Equal(common.Proposing, proposal.status)
	assert.Equal(common.Version(1), proposal.version)
	assert.NotNil(proposal.propoasl)
}

func Test_prepareConsensus(t *testing.T) {
	assert := assert.New(t)
	sp, _ := NewSoloPolicy(nil)
	proposal := mock_solo_proposal()

	err := sp.prepareConsensus(proposal)
	assert.NotNil(err)

	proposal.version = 1
	err = sp.prepareConsensus(proposal)
	assert.Nil(err)
	assert.Equal(common.Propose, proposal.status)
}

func Test_submitConsensus(t *testing.T) {
	assert := assert.New(t)
	proposal := mock_solo_proposal()
	sp, _ := NewSoloPolicy(nil)
	ok, err := sp.submitConsensus(proposal)
	assert.False(ok)
	assert.NotNil(err)

	proposal.status = common.Propose
	ok, err = sp.submitConsensus(proposal)
	assert.True(ok)
	assert.Nil(err)
	assert.Equal(common.Commited, proposal.status)
}

func Test_ToConsensus(t *testing.T) {
	assert := assert.New(t)
	proposal := mock_proposal()
	sp, _ := NewSoloPolicy(nil)
	ok, err := sp.ToConsensus(proposal)
	assert.True(ok)
	assert.Nil(err)
	assert.Equal(common.Version(1), version)
}
