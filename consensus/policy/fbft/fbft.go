package fbft

import (
	"errors"
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/config"
	"github.com/DSiSc/galaxy/consensus/messages"
	"github.com/DSiSc/galaxy/consensus/utils"
	"github.com/DSiSc/validator/tools/account"
	"math/rand"
	"sync/atomic"
	"time"
)

type FBFTPolicy struct {
	local account.Account
	name  string
	core  *fbftCore
	// time to reach consensus
	timeout      config.ConsensusTimeout
	isProcessing int32
}

func NewFBFTPolicy(timeout config.ConsensusTimeout, blockSwitch chan<- interface{}, emptyBlock bool, signVerify config.SignatureVerifySwitch) (*FBFTPolicy, error) {
	policy := &FBFTPolicy{
		name:    common.FbftPolicy,
		timeout: timeout,
	}
	policy.core = NewFBFTCore(blockSwitch, timeout, emptyBlock, signVerify)
	return policy, nil
}

func (instance *FBFTPolicy) Initialization(local account.Account, master account.Account, peers []account.Account, events types.EventCenter, onLine bool) {
	instance.local = local
	instance.core.nodes.Store(&nodesInfo{
		local:  local,
		master: master,
		peers:  peers,
	})
	instance.core.eventCenter = events
	instance.core.tolerance.Store(uint8((len(peers) - 1) / 3))
	log.Debug("start timeout master with view num %d.", instance.core.viewChange.GetCurrentViewNum())
	if !onLine && local.Address != master.Address {
		randDuration := rand.Int63n(5) * 1000
		go instance.core.waitMasterTimeout(time.Duration(instance.timeout.TimeoutToChangeView+randDuration) * time.Millisecond)
	}
}

func (instance *FBFTPolicy) PolicyName() string {
	return instance.name
}

func (instance *FBFTPolicy) Prepare(account account.Account) {
	instance.local = account

	originNodes := instance.core.nodes.Load().(*nodesInfo)
	newNodes := originNodes.clone()
	newNodes.local = account
	instance.core.nodes.Store(newNodes)
}

func (instance *FBFTPolicy) Start() {
	instance.core.Start()
}

func (instance *FBFTPolicy) ToConsensus(p *common.Proposal) error {
	if !atomic.CompareAndSwapInt32(&instance.isProcessing, 0, 1) {
		log.Warn("previous round have not finished")
		return errors.New("previous round have not finished")
	}
	defer atomic.StoreInt32(&instance.isProcessing, 0)
	var err error
	var result bool
	request := &messages.Request{
		Account:   instance.local,
		Timestamp: p.Timestamp,
		Payload:   p.Block,
	}
	timeToCollectResponseMsg := time.NewTimer(time.Duration(instance.timeout.TimeoutToWaitCommitMsg) * time.Millisecond)
	utils.SendEvent(instance.core, request)
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
		timeToCollectResponseMsg.Stop()
		instance.core.tryToCommit(p.Block, result)
		return err
	case <-timeToCollectResponseMsg.C:
		log.Error("consensus for digest %x timeout in %d seconds.", p.Block.Header.MixDigest, instance.timeout.TimeoutToCollectResponseMsg)
		instance.core.tryToCommit(p.Block, false)
		return fmt.Errorf("timeout for consensus")
	}
}

func (instance *FBFTPolicy) Halt() {
	return
}

func (instance *FBFTPolicy) GetConsensusResult() common.ConsensusResult {
	nodes := instance.core.nodes.Load().(*nodesInfo)
	log.Debug("now local is %d.", nodes.local.Extension.Id)
	return common.ConsensusResult{
		View:        instance.core.viewChange.GetCurrentViewNum(),
		Participate: nodes.peers,
		Master:      nodes.master,
	}
}

func (instance *FBFTPolicy) Online() {
	instance.core.sendOnlineRequest()
}
