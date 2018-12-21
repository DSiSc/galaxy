package fbft

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/config"
	"github.com/DSiSc/galaxy/consensus/messages"
	"github.com/DSiSc/galaxy/consensus/utils"
	"github.com/DSiSc/validator/tools/account"
	"time"
)

type FBFTPolicy struct {
	local account.Account
	name  string
	core  *fbftCore
	// time to reach consensus
	timeout config.ConsensusTimeout
}

func NewFBFTPolicy(account account.Account, timeout config.ConsensusTimeout, blockSwitch chan<- interface{}) (*FBFTPolicy, error) {
	policy := &FBFTPolicy{
		local:   account,
		name:    common.FbftPolicy,
		timeout: timeout,
	}
	policy.core = NewFBFTCore(account, blockSwitch)
	return policy, nil
}

func (instance *FBFTPolicy) Initialization(master account.Account, peers []account.Account, events types.EventCenter, onLine bool) {
	instance.core.nodes.master = master
	instance.core.nodes.peers = peers
	instance.core.eventCenter = events
	instance.core.tolerance = uint8((len(peers) - 1) / 3)
	instance.core.coreTimer = coreTimeout{
		timeToChangeViewTime:     instance.timeout.TimeoutToChangeView,
		timeToWaitCommitMsg:      instance.timeout.TimeoutToWaitCommitMsg,
		timeToCollectResponseMsg: instance.timeout.TimeoutToCollectResponseMsg,
	}
	log.Debug("start timeout master with view num %d.", instance.core.viewChange.GetCurrentViewNum())
	if nil != instance.core.coreTimer.timeToChangeViewTimer {
		instance.core.coreTimer.timeToChangeViewTimer.Stop()
	}
	if !onLine {
		instance.core.coreTimer.timeToChangeViewTimer = time.NewTimer(time.Duration(instance.timeout.TimeoutToChangeView) * time.Millisecond)
		go instance.core.waitMasterTimeout()
	}
}

func (instance *FBFTPolicy) PolicyName() string {
	return instance.name
}

func (instance *FBFTPolicy) Start() {
	log.Info("start fbft policy service.")
	instance.core.Start()
}

func (instance *FBFTPolicy) ToConsensus(p *common.Proposal) error {
	var err error
	var result bool
	request := &messages.Request{
		Account:   instance.local,
		Timestamp: p.Timestamp,
		Payload:   p.Block,
	}
	timeToCollectResponseMsg := time.NewTimer(time.Duration(instance.timeout.TimeoutToCollectResponseMsg) * time.Millisecond)
	go utils.SendEvent(instance.core, request)
	select {
	case consensusResult := <-instance.core.result:
		if nil != consensusResult.Result {
			log.Error("consensus for %x failed with error %v.", p.Block.Header.MixDigest, consensusResult.Result)
			err = consensusResult.Result
		} else {
			p.Block.Header.SigData = consensusResult.Signatures
			p.Block.HeaderHash = common.HeaderHash(p.Block)
			result = true
			log.Info("consensus for %x successfully with signature %x.", p.Block.Header.MixDigest, consensusResult.Signatures)
		}
		instance.core.commit(p.Block, result)
	case <-timeToCollectResponseMsg.C:
		log.Error("consensus for %x timeout in %d seconds.", p.Block.Header.MixDigest, instance.timeout)
		err = fmt.Errorf("timeout for consensus")
		instance.core.commit(p.Block, false)
	}
	return err
}

func (instance *FBFTPolicy) Halt() {
	return
}

func (instance *FBFTPolicy) GetConsensusResult() common.ConsensusResult {
	log.Debug("now local is %d.", instance.core.nodes.local.Extension.Id)
	return common.ConsensusResult{
		View:        instance.core.viewChange.GetCurrentViewNum(),
		Participate: instance.core.nodes.peers,
		Master:      instance.core.nodes.master,
	}
}

func (instance *FBFTPolicy) Online() {
	instance.core.sendOnlineRequest()
}
