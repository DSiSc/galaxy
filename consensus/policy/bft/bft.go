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
	policy := &BFTPolicy{
		name:         common.BFT_POLICY,
		participates: participates,
	}
	return policy, nil
}

func (self *BFTPolicy) PolicyName() string {
	return self.name
}
