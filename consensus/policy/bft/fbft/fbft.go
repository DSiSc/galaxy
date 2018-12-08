package fbft

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/policy/bft/messages"
	"github.com/DSiSc/galaxy/consensus/policy/bft/tools"
	commonr "github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/validator/tools/account"
	"time"
)

type FBFTPolicy struct {
	name string
	core *fbftCore
	// time to reach consensus
	timeout time.Duration
}

func NewFBFTPolicy(account account.Account, timeout int64, blockSwitch chan<- interface{}) (*FBFTPolicy, error) {
	policy := &FBFTPolicy{
		name:    common.FBFT_POLICY,
		timeout: time.Duration(timeout),
	}
	policy.core = NewFBFTCore(account, blockSwitch)
	return policy, nil
}

func (instance *FBFTPolicy) Initialization(role map[account.Account]commonr.Roler, peers []account.Account, events types.EventCenter) error {
	var masterExist = false
	for delegate, role := range role {
		if commonr.Master == role {
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
	return nil
}

func (instance *FBFTPolicy) PolicyName() string {
	return instance.name
}

func (instance *FBFTPolicy) Start() {
	log.Info("start fbft policy service.")
	instance.core.Start()
}

func (instance *FBFTPolicy) commit(block *types.Block, result bool) {
	commit := &messages.Commit{
		Account:    instance.core.nodes.local,
		Timestamp:  time.Now().Unix(),
		Digest:     block.Header.MixDigest,
		Signatures: block.Header.SigData,
		BlockHash:  block.HeaderHash,
		Result:     result,
	}
	instance.core.SendCommit(commit, block)
}

func (instance *FBFTPolicy) ToConsensus(p *common.Proposal) error {
	var err error
	var result bool
	request := &messages.Request{
		Timestamp: p.Timestamp,
		Payload:   p.Block,
	}
	timer := time.NewTimer(time.Second * instance.timeout)
	go tools.SendEvent(instance.core, request)
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
		instance.commit(p.Block, result)
	case <-timer.C:
		log.Error("consensus for %x timeout in %d seconds.", p.Block.Header.MixDigest, instance.timeout)
		err = fmt.Errorf("timeout for consensus")
		result = false
		instance.commit(p.Block, result)
	}
	return err
}

func (instance *FBFTPolicy) Halt() {
	return
}

func (self *FBFTPolicy) GetConsensusResult() common.ConsensusResult {
	role := make(map[account.Account]commonr.Roler)
	for _, peer := range self.core.nodes.peers {
		if self.core.nodes.master == peer {
			role[peer] = commonr.Master
			log.Debug("now master is %d.", peer.Extension.Id)
			continue
		}
		role[peer] = commonr.Slave
	}
	log.Debug("now local is %d.", self.core.nodes.local.Extension.Id)
	return common.ConsensusResult{
		View:        uint64(0),
		Participate: self.core.nodes.peers,
		Roles:       role,
	}
}
