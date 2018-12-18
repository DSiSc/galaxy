package bft

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/messages"
	"github.com/DSiSc/galaxy/consensus/utils"
	commonr "github.com/DSiSc/galaxy/role/common"
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

func NewBFTPolicy(account account.Account, timeout int64) (*BFTPolicy, error) {
	policy := &BFTPolicy{
		name:    common.BFT_POLICY,
		account: account,
		timeout: time.Duration(timeout),
		result:  make(chan *messages.ConsensusResult),
	}
	policy.bftCore = NewBFTCore(account, policy.result)
	return policy, nil
}

func (self *BFTPolicy) Initialization(role map[account.Account]commonr.Roler, peers []account.Account, events types.EventCenter, onLine bool) error {
	if onLine{
		log.Info("online first time.")
	}
	if len(role) != len(peers) {
		log.Error("bft core has not been initial, please confirm.")
		return fmt.Errorf("role and peers not in consistent")
	}
	var masterExist bool = false
	for delegate, role := range role {
		if commonr.Master == role {
			self.bftCore.master = delegate.Extension.Id
			masterExist = true
			break
		}
	}
	if !masterExist {
		log.Error("no master exist in delegates")
		return fmt.Errorf("no master")
	}
	self.bftCore.commit = false
	self.bftCore.peers = peers
	self.bftCore.eventCenter = events
	self.bftCore.tolerance = uint8((len(peers) - 1) / 3)
	self.bftCore.signature = &signData{
		signatures: make([][]byte, 0),
		signMap:    make(map[account.Account][]byte),
	}
	return nil
}

func (self *BFTPolicy) PolicyName() string {
	return self.name
}

func (self *BFTPolicy) Start() {
	log.Info("start bft policy service.")
	self.bftCore.Start(self.account)
}

func (self *BFTPolicy) commit(block *types.Block, result bool) {
	commit := &messages.Commit{
		Account:    self.account,
		Timestamp:  time.Now().Unix(),
		Digest:     block.Header.MixDigest,
		Signatures: block.Header.SigData,
		BlockHash:  block.HeaderHash,
		Result:     result,
	}
	self.bftCore.SendCommit(commit, block)
}

func (self *BFTPolicy) ToConsensus(p *common.Proposal) error {
	var err error
	var result = false
	request := &messages.Request{
		Timestamp: p.Timestamp,
		Payload:   p.Block,
	}
	timer := time.NewTimer(time.Second * self.timeout)
	go utils.SendEvent(self.bftCore, request)
	select {
	case consensusResult := <-self.result:
		if nil != consensusResult.Result {
			log.Error("consensus for %x failed with error %v.", p.Block.Header.MixDigest, consensusResult.Result)
			err = consensusResult.Result
		} else {
			p.Block.Header.SigData = consensusResult.Signatures
			p.Block.HeaderHash = common.HeaderHash(p.Block)
			result = true
			log.Info("consensus for %x successfully with signature %x.", p.Block.Header.MixDigest, consensusResult.Signatures)
		}
		go self.commit(p.Block, result)
	case <-timer.C:
		log.Error("consensus for %x timeout in %d seconds.", p.Block.Header.MixDigest, self.timeout)
		err = fmt.Errorf("timeout for consensus")
		go self.commit(p.Block, result)
	}
	return err
}

func (self *BFTPolicy) Halt() {
	return
}

func (self *BFTPolicy) GetConsensusResult() common.ConsensusResult {
	return common.ConsensusResult{
		View:        uint64(0),
		Participate: self.bftCore.peers,
		Roles:       make(map[account.Account]commonr.Roler),
	}
}

func (self *BFTPolicy) Online() {
	return
}
