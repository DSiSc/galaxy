package solo

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	commonr "github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/validator"
	"github.com/DSiSc/validator/tools/account"
	"github.com/DSiSc/validator/tools/signature"
	"math"
)

type SoloPolicy struct {
	name        string
	account     account.Account
	tolerance   uint8
	version     common.Version
	peers       []account.Account
	role        map[account.Account]commonr.Roler
	receipts    types.Receipts
	eventCenter types.EventCenter
	blockSwitch chan<- interface{}
}

// SoloProposal that with solo policy
type SoloProposal struct {
	proposal *common.Proposal
	version  common.Version
	status   common.ConsensusStatus
}

func NewSoloPolicy(account account.Account, blkSwitch chan<- interface{}) (*SoloPolicy, error) {
	policy := &SoloPolicy{
		name:        common.SOLO_POLICY,
		account:     account,
		tolerance:   common.SOLO_CONSENSUS_NUM,
		blockSwitch: blkSwitch,
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

func (self *SoloPolicy) Initialization(role map[account.Account]commonr.Roler, account []account.Account, event types.EventCenter) error {
	log.Info("Initial solo policy.")
	if len(role) != len(account) {
		log.Error("solo core has not been initial, please confirm.")
		return fmt.Errorf("role and peers not in consistent")
	}
	if common.SOLO_CONSENSUS_NUM != uint8(len(account)) {
		return fmt.Errorf("solo policy only support one participate")
	}
	self.role = role
	self.peers = account
	self.eventCenter = event
	return nil
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
	var err error
	proposal := self.toSoloProposal(p)
	err = self.prepareConsensus(proposal)
	if err != nil {
		log.Error("Prepare proposal failed.")
		self.eventCenter.Notify(types.EventConsensusFailed, nil)
		return fmt.Errorf("prepare proposal failed")
	}
	ok := self.toConsensus(proposal)
	if ok == false {
		log.Error("Local verify failed.")
		self.eventCenter.Notify(types.EventConsensusFailed, nil)
		return fmt.Errorf("local verify failed")
	}
	// verify num of sign
	signData := proposal.proposal.Block.Header.SigData
	var validSign = make(map[types.Address][]byte)
	for _, value := range signData {
		signAddress, err := signature.Verify(p.Block.Header.MixDigest, value)
		if err != nil {
			log.Error("Invalid signature is %x.", value)
			continue
		}
		validSign[signAddress] = value
	}
	if uint8(len(validSign)) < common.SOLO_CONSENSUS_NUM {
		log.Error("Not enough valid signature which is %d.", len(validSign))
		self.eventCenter.Notify(types.EventConsensusFailed, nil)
		return fmt.Errorf("not enough valid signature")
	}
	if _, ok := validSign[self.account.Address]; !ok {
		log.Error("absence self signature.")
		self.eventCenter.Notify(types.EventConsensusFailed, nil)
		return fmt.Errorf("absence self signature")
	}
	err = self.submitConsensus(proposal)
	if err != nil {
		log.Error("Submit proposal failed.")
		self.eventCenter.Notify(types.EventConsensusFailed, nil)
		return fmt.Errorf("submit proposal failed")
	}
	if proposal.status != common.Committed {
		log.Error("Not to consensus.")
		self.eventCenter.Notify(types.EventConsensusFailed, nil)
		return fmt.Errorf("consensus status fault")
	}
	self.version = proposal.version
	p.Block.HeaderHash = common.HeaderHash(p.Block)
	self.commitBlock(p.Block)
	return err
}

func (self *SoloPolicy) commitBlock(block *types.Block) {
	log.Info("send block to block switch.")
	self.blockSwitch <- block
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
		log.Error("Proposal status must be Propose to submit consensus.")
		return fmt.Errorf("proposal status must be Propose")
	}
	p.status = common.Committed
	return nil
}

func (self *SoloPolicy) toConsensus(p *SoloProposal) bool {
	if uint8(len(self.peers)) != common.SOLO_CONSENSUS_NUM {
		log.Error("Solo participates invalid.")
		return false
	}

	local := self.peers[0]
	validators := validator.NewValidator(&local)
	_, ok := validators.ValidateBlock(p.proposal.Block)
	if nil != ok {
		log.Error("Validator verify failed.")
		return false
	}
	log.Info("Consensus reached for block %d.", p.proposal.Block.Header.Height)
	self.receipts = validators.Receipts
	return true
}

func (self *SoloPolicy) GetConsensusResult() common.ConsensusResult {
	return common.ConsensusResult{
		View:        uint64(0),
		Participate: self.peers,
		Roles:       self.role,
	}
}
