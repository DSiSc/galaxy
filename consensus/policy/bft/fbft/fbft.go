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
	"math"
	"time"
)

type FBFTPolicy struct {
	name    string
	account account.Account
	core    *fbftCore
	timeout time.Duration
	result  chan *messages.ConsensusResult
}

func NewFBFTPolicy(account account.Account, timeout int64, blockSwitch chan<- interface{}) (*FBFTPolicy, error) {
	policy := &FBFTPolicy{
		name:    common.FBFT_POLICY,
		account: account,
		timeout: time.Duration(timeout),
		result:  make(chan *messages.ConsensusResult),
	}
	policy.core = NewFBFTCore(account, policy.result, blockSwitch)
	return policy, nil
}

func (instance *FBFTPolicy) Initialization(role map[account.Account]commonr.Roler, peers []account.Account, events types.EventCenter) error {
	instance.core.master = math.MaxUint64
	for delegate, role := range role {
		if commonr.Master == role {
			instance.core.master = delegate.Extension.Id
		}
	}
	if instance.core.master == math.MaxUint64 {
		log.Error("no master exist in delegates")
		return fmt.Errorf("no master")
	}
	instance.core.commit = false
	instance.core.peers = peers
	instance.core.eventCenter = events
	instance.core.tolerance = uint8((len(peers) - 1) / 3)
	instance.core.signature = &tools.SignData{
		Signatures: make([][]byte, 0),
		SignMap:    make(map[account.Account][]byte),
	}
	return nil
}

func (instance *FBFTPolicy) PolicyName() string {
	return instance.name
}

func (instance *FBFTPolicy) Start() {
	log.Info("start fbft policy service.")
	instance.core.Start(instance.account)
}

func (instance *FBFTPolicy) commit(block *types.Block, result bool) {
	commit := &messages.Commit{
		Account:    instance.account,
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
	case consensusResult := <-instance.result:
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
	for _, peer := range self.core.peers {
		if self.core.master == peer.Extension.Id {
			role[peer] = commonr.Master
			log.Warn("now master is %d.", peer.Extension.Id)
			continue
		}
		role[peer] = commonr.Slave
	}
	log.Warn("now local is %d.", self.core.local.Extension.Id)
	return common.ConsensusResult{
		View:        uint64(0),
		Participate: self.core.peers,
		Roles:       role,
	}
}
