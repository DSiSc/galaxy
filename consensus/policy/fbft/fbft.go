package fbft

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/messages"
	"github.com/DSiSc/galaxy/consensus/utils"
	roleCommon "github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/validator/tools/account"
	"time"
)

type FBFTPolicy struct {
	local account.Account
	name  string
	core  *fbftCore
	// time to reach consensus
	timeout time.Duration
}

func NewFBFTPolicy(account account.Account, timeout int64, blockSwitch chan<- interface{}) (*FBFTPolicy, error) {
	policy := &FBFTPolicy{
		local:   account,
		name:    common.FBFT_POLICY,
		timeout: time.Duration(timeout),
	}
	policy.core = NewFBFTCore(account, blockSwitch)
	return policy, nil
}

func (instance *FBFTPolicy) Initialization(role map[account.Account]roleCommon.Roler, peers []account.Account, events types.EventCenter) error {
	var masterExist = false
	for delegate, role := range role {
		if roleCommon.Master == role {
			instance.core.nodes.master = delegate
			masterExist = true
		}
	}
	if !masterExist {
		log.Error("no master exist, please confirm.")
		return fmt.Errorf("no master exist")
	}
	instance.core.nodes.peers = peers
	instance.core.eventCenter = events
	instance.core.tolerance = uint8((len(peers) - 1) / 3)
	log.Info("start timeout master with view num %d.", instance.core.viewChange.GetCurrentViewNum())
	if nil != instance.core.timeoutTimer {
		instance.core.timeoutTimer.Stop()
	}
	instance.core.timeoutTimer = time.NewTimer(30 * time.Second)
	go instance.core.waitMasterTimeout()
	return nil
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
	timer := time.NewTimer(time.Second * instance.timeout)
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
	case <-timer.C:
		log.Error("consensus for %x timeout in %d seconds.", p.Block.Header.MixDigest, instance.timeout)
		err = fmt.Errorf("timeout for consensus")
		result = false
		instance.core.commit(p.Block, result)
	}
	return err
}

func (instance *FBFTPolicy) Halt() {
	return
}

func (instance *FBFTPolicy) GetConsensusResult() common.ConsensusResult {
	role := make(map[account.Account]roleCommon.Roler)
	for _, peer := range instance.core.nodes.peers {
		if instance.core.nodes.master == peer {
			role[peer] = roleCommon.Master
			log.Debug("now master is %d.", peer.Extension.Id)
			continue
		}
		role[peer] = roleCommon.Slave
	}
	log.Debug("now local is %d.", instance.core.nodes.local.Extension.Id)
	return common.ConsensusResult{
		View:        uint64(0),
		Participate: instance.core.nodes.peers,
		Roles:       role,
	}
}

func (instance *FBFTPolicy) Online() {
	instance.core.sendOnlineRequest()
}
