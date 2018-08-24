package policy

import (
	"fmt"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/txpool/log"
)

var version common.Version

const (
	POLICY_NAME = "solo"
)

type SoloPolicy struct {
	name         string
	participates participates.Participates
}

// SoloProposal that with solo policy
type SoloProposal struct {
	propoasl *common.Proposal
	version  common.Version
	status   common.ConsensusStatus
}

func NewSoloPolicy(participates participates.Participates) (*SoloPolicy, error) {
	policy := &SoloPolicy{
		name:         POLICY_NAME,
		participates: participates,
	}
	version = 0
	return policy, nil
}

func (self *SoloPolicy) PolicyName() string {
	return self.name
}

func toSoloProposal(p *common.Proposal) *SoloProposal {
	return &SoloProposal{
		propoasl: p,
		version:  version + 1,
		status:   common.Proposing,
	}
}

// to get consensus
func (self *SoloPolicy) ToConsensus(p *common.Proposal) (bool, error) {
	if p.Block == nil {
		log.Error("Block segment cant not be nil in proposal.")
		return false, fmt.Errorf("Proposal segment fault.")
	}

	proposal := toSoloProposal(p)
	// prepare
	err := self.prepareConsensus(proposal)
	if err != nil {
		log.Error("Prepare proposal failed.")
		return false, fmt.Errorf("Prepare proposal failed.")
	}
	// proposal
	state, err := self.submitConsensus(proposal)
	if err != nil {
		log.Error("Sunmit proposal failed.")
		return false, fmt.Errorf("Sunmit proposal failed.")
	}

	if proposal.status != common.Commited {
		log.Error("Not to consensus.")
		return false, fmt.Errorf("Not to consensus..")
	}
	version = proposal.version
	return state, nil
}

// check proposal param and set consensus status
func (self *SoloPolicy) prepareConsensus(p *SoloProposal) error {
	if p.version <= version {
		log.Error("Proposal version segment less than version which has configmed.")
		return fmt.Errorf("Proposal version less than confirmed.")
	}
	if p.status != common.Proposing {
		log.Error("Proposal status must be Proposal befor submit consensus.")
		return fmt.Errorf("Proposal status must be Proposal.")
	}
	p.status = common.Propose
	return nil
}

func (self *SoloPolicy) submitConsensus(p *SoloProposal) (bool, error) {
	if p.status != common.Propose {
		log.Error("Proposal status must be Proposaling to submit consensus.")
		return false, fmt.Errorf("Proposal status must be Proposaling.")
	}
	// TODO: collect result of every participates
	p.status = common.Commited
	return true, nil
}
