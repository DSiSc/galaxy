package solo

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/config"
	"github.com/DSiSc/validator"
	"github.com/DSiSc/validator/tools/account"
	"github.com/DSiSc/validator/tools/signature"
	"math"
)

type SoloPolicy struct {
	name                       string
	local                      account.Account
	tolerance                  uint8
	version                    common.Version
	peers                      []account.Account
	master                     account.Account
	eventCenter                types.EventCenter
	blockSwitch                chan<- interface{}
	enableEmptyBlock           bool
	enableLocalSignatureVerify bool
}

// SoloProposal that with solo policy
type SoloProposal struct {
	proposal *common.Proposal
	version  common.Version
	status   common.ConsensusStatus
}

func NewSoloPolicy(account account.Account, blkSwitch chan<- interface{}, enable bool, signVerify config.SignatureVerifySwitch) (*SoloPolicy, error) {
	policy := &SoloPolicy{
		name:                       common.SoloPolicy,
		local:                      account,
		tolerance:                  common.SoloConsensusNum,
		blockSwitch:                blkSwitch,
		enableEmptyBlock:           enable,
		enableLocalSignatureVerify: signVerify.LocalVerifySignature,
	}
	return policy, nil
}

func (instance *SoloPolicy) PolicyName() string {
	return instance.name
}

func (instance *SoloPolicy) Start() {
	log.Info("Start solo policy service.")
	return
}

func (instance *SoloPolicy) Halt() {
	log.Warn("Stop solo policy service.")
	return
}

func (instance *SoloPolicy) Initialization(master account.Account, accounts []account.Account, event types.EventCenter, onLine bool) {
	if onLine {
		log.Debug("online first time.")
	}
	if common.SoloConsensusNum != uint8(len(accounts)) {
		panic("solo policy only support one participate")
	}
	instance.master = master
	instance.peers = accounts
	instance.eventCenter = event
	return
}

func (instance *SoloPolicy) toSoloProposal(p *common.Proposal) *SoloProposal {
	if instance.version == math.MaxUint64 {
		instance.version = 0
	}
	return &SoloProposal{
		proposal: p,
		version:  instance.version + 1,
		status:   common.Proposing,
	}
}

// to get consensus
func (instance *SoloPolicy) ToConsensus(p *common.Proposal) error {
	var err error
	proposal := instance.toSoloProposal(p)
	err = instance.prepareConsensus(proposal)
	if err != nil {
		log.Error("Prepare proposal failed.")
		instance.eventCenter.Notify(types.EventConsensusFailed, nil)
		return fmt.Errorf("prepare proposal failed")
	}
	ok := instance.toConsensus(proposal)
	if ok == false {
		log.Error("Local verify failed.")
		instance.eventCenter.Notify(types.EventConsensusFailed, nil)
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
	if uint8(len(validSign)) < common.SoloConsensusNum {
		log.Error("Not enough valid signature which is %d.", len(validSign))
		instance.eventCenter.Notify(types.EventConsensusFailed, nil)
		return fmt.Errorf("not enough valid signature")
	}
	if _, ok := validSign[instance.local.Address]; !ok {
		log.Error("absence self signature.")
		instance.eventCenter.Notify(types.EventConsensusFailed, nil)
		return fmt.Errorf("absence self signature")
	}
	err = instance.submitConsensus(proposal)
	if err != nil {
		log.Error("Submit proposal failed.")
		instance.eventCenter.Notify(types.EventConsensusFailed, nil)
		return fmt.Errorf("submit proposal failed")
	}
	if proposal.status != common.Committed {
		log.Error("Not to consensus.")
		instance.eventCenter.Notify(types.EventConsensusFailed, nil)
		return fmt.Errorf("consensus status fault")
	}
	instance.version = proposal.version
	p.Block.HeaderHash = common.HeaderHash(p.Block)
	instance.commitBlock(p.Block)
	return err
}

func (instance *SoloPolicy) commitBlock(block *types.Block) {
	log.Info("send block to block switch.")
	if !instance.enableEmptyBlock && 0 == len(block.Transactions) {
		log.Warn("block without transaction.")
		instance.eventCenter.Notify(types.EventBlockWithoutTxs, nil)
		return
	}
	instance.blockSwitch <- block
}

func (instance *SoloPolicy) prepareConsensus(p *SoloProposal) error {
	if p.version <= instance.version {
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

func (instance *SoloPolicy) submitConsensus(p *SoloProposal) error {
	if p.status != common.Propose {
		log.Error("Proposal status must be Propose to submit consensus.")
		return fmt.Errorf("proposal status must be Propose")
	}
	p.status = common.Committed
	return nil
}

func (instance *SoloPolicy) toConsensus(p *SoloProposal) bool {
	if uint8(len(instance.peers)) != common.SoloConsensusNum {
		log.Error("Solo participates invalid.")
		return false
	}

	local := instance.peers[0]
	validators := validator.NewValidator(&local)
	_, ok := validators.ValidateBlock(p.proposal.Block, instance.enableLocalSignatureVerify)
	if nil != ok {
		log.Error("Validator verify failed.")
		return false
	}
	log.Info("Consensus reached for block %d.", p.proposal.Block.Header.Height)
	return true
}

func (instance *SoloPolicy) GetConsensusResult() common.ConsensusResult {
	return common.ConsensusResult{
		View:        uint64(0),
		Participate: instance.peers,
		Master:      instance.master,
	}
}

func (instance *SoloPolicy) Online() {
	instance.eventCenter.Notify(types.EventOnline, nil)
	return
}
