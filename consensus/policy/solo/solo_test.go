package solo

import (
	"fmt"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/participates/config"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func Test_NewSoloPolicy(t *testing.T) {
	asserts := assert.New(t)
	sp, err := NewSoloPolicy(nil)
	asserts.Nil(err)
	asserts.NotNil(sp)
	asserts.Equal(common.SOLO_POLICY, sp.name)
	asserts.Nil(sp.participates)
	asserts.Equal(common.SOLO_POLICY, sp.PolicyName())
}

func mock_proposal() *common.Proposal {
	var block types.Block
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

var MockParticipate, _ = participates.NewParticipates(config.ParticipateConfig{
	PolicyName: common.SOLO_POLICY,
})

func Test_toSoloProposal(t *testing.T) {
	asserts := assert.New(t)
	p := mock_proposal()
	sp, _ := NewSoloPolicy(MockParticipate)
	proposal := sp.toSoloProposal(p)
	asserts.NotNil(proposal)
	asserts.Equal(common.Proposing, proposal.status)
	asserts.Equal(common.Version(1), proposal.version)
	asserts.NotNil(proposal.propoasl)
}

func Test_prepareConsensus(t *testing.T) {
	asserts := assert.New(t)
	sp, _ := NewSoloPolicy(MockParticipate)
	proposal := mock_solo_proposal()

	err := sp.prepareConsensus(proposal)
	asserts.NotNil(err)

	proposal.version = 1
	err = sp.prepareConsensus(proposal)
	asserts.Nil(err)
	asserts.Equal(common.Propose, proposal.status)

	proposal.status = common.Propose
	err = sp.prepareConsensus(proposal)
	asserts.NotNil(err)

	sp.version = math.MaxUint64
	proposal = sp.toSoloProposal(nil)
	err = sp.prepareConsensus(proposal)
	asserts.Nil(err)
}

func Test_submitConsensus(t *testing.T) {
	asserts := assert.New(t)
	proposal := mock_solo_proposal()
	sp, _ := NewSoloPolicy(MockParticipate)
	err := sp.submitConsensus(proposal)
	asserts.NotNil(err)

	proposal.status = common.Propose
	err = sp.submitConsensus(proposal)
	asserts.Nil(err)
	asserts.Equal(common.Committed, proposal.status)
}

func TestSoloPolicy_ToConsensus(t *testing.T) {
	asserts := assert.New(t)
	proposal := mock_proposal()
	sp, _ := NewSoloPolicy(MockParticipate)

	err := sp.ToConsensus(proposal)
	asserts.NotNil(err)
	asserts.Equal(common.Version(0), sp.version)

	proposal.Block = nil
	err = sp.ToConsensus(proposal)
	asserts.NotNil(err)
	excErr := fmt.Errorf("proposal block is nil")
	asserts.Equal(excErr, err)
}

func TestPrepareConsensus(t *testing.T) {
	asserts := assert.New(t)
	sp, _ := NewSoloPolicy(MockParticipate)
	sp.version = math.MaxUint64
	proposal := sp.toSoloProposal(nil)
	err := sp.prepareConsensus(proposal)
	asserts.Nil(err)
}

func TestNewSoloPolicy(t *testing.T) {
	sp, err := NewSoloPolicy(MockParticipate)
	assert.Nil(t, err)
	assert.NotNil(t, sp)
	assert.Equal(t, common.SOLO_POLICY, sp.name)
	assert.Equal(t, common.Version(0), sp.version)
}

func TestSoloPolicy_PolicyName(t *testing.T) {
	sp, err := NewSoloPolicy(MockParticipate)
	assert.Nil(t, err)
	assert.NotNil(t, sp)
	assert.Equal(t, common.SOLO_POLICY, sp.PolicyName())
}

func Test_toConsensus(t *testing.T) {
	sp, _ := NewSoloPolicy(MockParticipate)
	err := sp.toConsensus(nil)
	assert.Equal(t, false, err)
}
