package solo

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/validator"
	"github.com/DSiSc/validator/tools/signature"
	"math"
)

type SoloPolicy struct {
	name         string
	version      common.Version
	participates participates.Participates
	tolerance    uint8
}

// SoloProposal that with solo policy
type SoloProposal struct {
	proposal *common.Proposal
	version  common.Version
	status   common.ConsensusStatus
}

func NewSoloPolicy(participates participates.Participates) (*SoloPolicy, error) {
	policy := &SoloPolicy{
		name:         common.SOLO_POLICY,
		participates: participates,
		version:      0,
		tolerance:    uint8(1),
	}
	return policy, nil
}

func (self *SoloPolicy) PolicyName() string {
	return self.name
}

func (self *SoloPolicy) Start() {
	log.Info("Start solo policy service.")
	return
}

func (self *SoloPolicy) Halt() {
	log.Warn("Stop solo policy service.")
	return
}

func (self *SoloPolicy) toSoloProposal(p *common.Proposal) *SoloProposal {
	if self.version == math.MaxUint64 {
		self.version = 0
	}
	return &SoloProposal{
		proposal: p,
		version:  self.version + 1,
		status:   common.Proposing,
	}
}

// to get consensus
func (self *SoloPolicy) ToConsensus(p *common.Proposal) error {
	if p.Block == nil {
		log.Error("Block segment cant not be nil in proposal.")
		return fmt.Errorf("proposal block is nil")
	}
	// to issue proposal
	proposal := self.toSoloProposal(p)
	// prepare
	err := self.prepareConsensus(proposal)
	if err != nil {
		log.Error("Prepare proposal failed.")
		return fmt.Errorf("prepare proposal failed")
	}
	// get consensus
	ok := self.toConsensus(proposal)
	if ok == false {
		log.Error("Local verify failed.")
		return fmt.Errorf("local verify failed")
	}
	// verify num of sign
	signData := proposal.proposal.Block.Header.SigData
	if uint8(len(signData)) < self.tolerance {
		log.Error("Not enough signature.")
		return fmt.Errorf("not enough signature")
	} else {
		var validSign = make(map[types.Address][]byte)
		var signAddress types.Address
		for _, value := range signData {
			signAddress, err = signature.Verify(p.Block.Header.MixDigest, value)
			if err != nil {
				log.Error("Invalid signature is %x.", value)
				continue
			}
			validSign[signAddress] = value
		}
		if len(validSign) < common.SOLO_CONSENSUS_NUM {
			log.Error("Not enough valid signature which is %d.", len(validSign))
			return fmt.Errorf("not enough valid signature")
		}
	}
	// committed
	err = self.submitConsensus(proposal)
	if err != nil {
		log.Error("Submit proposal failed.")
		return fmt.Errorf("submit proposal failed")
	}
	// just a check
	if proposal.status != common.Committed {
		log.Error("Not to consensus.")
		return fmt.Errorf("consensus status fault")
	}
	self.version = proposal.version
	return nil
}

func (self *SoloPolicy) prepareConsensus(p *SoloProposal) error {
	if p.version <= self.version {
		log.Error("Proposal version segment less than version which has confirmed.")
		return fmt.Errorf("proposal version less than confirmed")
	}
	if p.status != common.Proposing {
		log.Error("Proposal status must be Proposal befor submit consensus.")
		return fmt.Errorf("proposal status must be in proposal")
	}
	p.status = common.Propose
	return nil
}

func (self *SoloPolicy) submitConsensus(p *SoloProposal) error {
	if p.status != common.Propose {
		log.Error("Proposal status must be Proposaling to submit consensus.")
		return fmt.Errorf("proposal status must be proposaling")
	}
	p.status = common.Committed
	return nil
}

func (self *SoloPolicy) toConsensus(p *SoloProposal) bool {
	if nil == p {
		log.Error("Proposal invalid.")
		return false
	}

	member, err := self.participates.GetParticipates()
	if len(member) != common.SOLO_CONSENSUS_NUM || err != nil {
		log.Error("Solo participates invalid.")
		return false
	}
	// SOLO, so we just verify it local
	local := member[0]
	validators := validator.NewValidator(&local)
	_, ok := validators.ValidateBlock(p.proposal.Block)
	if nil != ok {
		log.Error("Validator verify failed.")
		return false
	}
	log.Info("Consensus reached for block %d.", p.proposal.Block.Header.Height)
	return true
}
