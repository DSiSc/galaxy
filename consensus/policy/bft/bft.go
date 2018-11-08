package bft

import (
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/participates"
)

type BFTPolicy struct {
	name         string
	tolerance    uint8
	participates participates.Participates
}

// BFTProposal that with solo policy
type BFTProposal struct {
	proposal *common.Proposal
	status   common.ConsensusStatus
}

func NewBFTPolicy(participates participates.Participates) (*BFTPolicy, error) {
	participate, _ := participates.GetParticipates()
	policy := &BFTPolicy{
		name:         common.BFT_POLICY,
		participates: participates,
		tolerance:    uint8((len(participate) - 1) / 3),
	}
	return policy, nil
}

func (self *BFTPolicy) PolicyName() string {
	return self.name
}

// to get consensus
func (self *BFTPolicy) ToConsensus(p *common.Proposal) error {
	return nil
}
