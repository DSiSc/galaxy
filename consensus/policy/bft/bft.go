package bft

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

type BFTPolicy struct {
	name string
	// local account
	account account.Account
	bftCore *bftCore
	timeout time.Duration
	result  chan *messages.ConsensusResult
}

func NewBFTPolicy(account account.Account, timeout config.ConsensusTimeout) (*BFTPolicy, error) {
	policy := &BFTPolicy{
		name:    common.BftPolicy,
		account: account,
		timeout: time.Duration(timeout.TimeoutToChangeView),
		result:  make(chan *messages.ConsensusResult),
	}
	policy.bftCore = NewBFTCore(account, policy.result)
	return policy, nil
}

func (instance *BFTPolicy) Initialization(master account.Account, peers []account.Account, events types.EventCenter, onLine bool) {
	if onLine {
		log.Debug("online first time.")
	}
	instance.bftCore.master = master
	instance.bftCore.commit = false
	instance.bftCore.peers = peers
	instance.bftCore.eventCenter = events
	instance.bftCore.tolerance = uint8((len(peers) - 1) / 3)
	instance.bftCore.signature = &signData{
		signatures: make([][]byte, 0),
		signMap:    make(map[account.Account][]byte),
	}
	return
}

func (instance *BFTPolicy) PolicyName() string {
	return instance.name
}

func (instance *BFTPolicy) Start() {
	log.Info("start bft policy service.")
	instance.bftCore.Start(instance.account)
}

func (instance *BFTPolicy) commit(block *types.Block, result bool) {
	commit := &messages.Commit{
		Account:    instance.account,
		Timestamp:  time.Now().Unix(),
		Digest:     block.Header.MixDigest,
		Signatures: block.Header.SigData,
		BlockHash:  block.HeaderHash,
		Result:     result,
	}
	instance.bftCore.SendCommit(commit, block)
}

func (instance *BFTPolicy) ToConsensus(p *common.Proposal) error {
	var err error
	var result = false
	request := &messages.Request{
		Timestamp: p.Timestamp,
		Payload:   p.Block,
	}
	timer := time.NewTimer(time.Second * instance.timeout)
	go utils.SendEvent(instance.bftCore, request)
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
		go instance.commit(p.Block, result)
	case <-timer.C:
		log.Error("consensus for %x timeout in %d seconds.", p.Block.Header.MixDigest, instance.timeout)
		err = fmt.Errorf("timeout for consensus")
		go instance.commit(p.Block, result)
	}
	return err
}

func (instance *BFTPolicy) Halt() {
	return
}

func (instance *BFTPolicy) GetConsensusResult() common.ConsensusResult {
	return common.ConsensusResult{
		View:        uint64(0),
		Participate: instance.bftCore.peers,
		Master:      instance.bftCore.master,
	}
}

func (instance *BFTPolicy) Online() {
	return
}
