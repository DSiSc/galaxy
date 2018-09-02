package policy

import (
	"fmt"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/txpool/log"
	"github.com/DSiSc/validator"
)

var version common.Version

const (
	SOLO_POLICY   = "solo"
	CONSENSUS_NUM = 1
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
		name:         SOLO_POLICY,
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
func (self *SoloPolicy) ToConsensus(p *common.Proposal) error {
	if p.Block == nil {
		log.Error("Block segment cant not be nil in proposal.")
		return fmt.Errorf("Proposal segment fault.")
	}
	// to issue proposal
	proposal := toSoloProposal(p)
	// prepare
	err := self.prepareConsensus(proposal)
	if err != nil {
		log.Error("Prepare proposal failed.")
		return fmt.Errorf("Prepare proposal failed.")
	}
	// get consensus
	ok := self.toConsensus(proposal)
	if ok == false {
		log.Error("Local verify failed.")
		return fmt.Errorf("Local verify failed.")
	}
	// verify consensus result
	signData := proposal.propoasl.Block.Header.SigData
	if len(signData) >= CONSENSUS_NUM {
		log.Error("Not enough signature.")
		return fmt.Errorf("Not enough signature.")
	}
	// committed
	err = self.submitConsensus(proposal)
	if err != nil {
		log.Error("Sunmit proposal failed.")
		return fmt.Errorf("Sunmit proposal failed.")
	}
	// just a check
	if proposal.status != common.Committed {
		log.Error("Not to consensus.")
		return fmt.Errorf("Not to consensus.")
	}
	version = proposal.version
	return nil
}

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

func (self *SoloPolicy) submitConsensus(p *SoloProposal) error {
	if p.status != common.Propose {
		log.Error("Proposal status must be Proposaling to submit consensus.")
		return fmt.Errorf("Proposal status must be Proposaling.")
	}
	p.status = common.Committed
	return nil
}

func (self *SoloPolicy) toConsensus(p *SoloProposal) bool {
	if nil != p {
		log.Error("Proposal invalid.")
		return false
	}

	member, err := self.participates.GetParticipates()
	if len(member) != 1 || err != nil {
		log.Error("Solo participates invalid.")
		return false
	}
	// SOLO, so we just verify it local
	local := member[0]
	validators := validator.NewValidator(&local)
	_, ok := validators.ValidateBlock(p.propoasl.Block)
	if nil != ok {
		log.Error("Validator verify failed.")
		return false
	}
	log.Info("Validator verify success in consensus.")
	return true
}
