package bft

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/policy/bft/messages"
	"github.com/DSiSc/galaxy/consensus/policy/bft/tools"
	commonr "github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/validator/tools/account"
)

type BFTPolicy struct {
	name string
	// local account
	account account.Account
	bftCore *bftCore
	result  chan messages.SignatureSet
}

func NewBFTPolicy(account account.Account) (*BFTPolicy, error) {
	policy := &BFTPolicy{
		name:    common.BFT_POLICY,
		account: account,
		result:  make(chan messages.SignatureSet),
	}
	policy.bftCore = NewBFTCore(account, policy.result)
	return policy, nil
}

func (self *BFTPolicy) Initialization(role map[account.Account]commonr.Roler, peers []account.Account) error {
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
	self.bftCore.tolerance = uint8((len(peers) - 1) / 3)
	return nil
}

func (self *BFTPolicy) PolicyName() string {
	return self.name
}

func (self *BFTPolicy) Start() {
	log.Info("start bft policy service.")
	self.bftCore.Start(self.account)
}

func (self *BFTPolicy) toConsensus(result messages.SignatureSet) error {
	return nil
}

func (self *BFTPolicy) ToConsensus(p *common.Proposal) error {
	request := &messages.Request{
		Timestamp: p.Timestamp,
		Payload:   p.Block,
	}
	go tools.SendEvent(self.bftCore, request)
	result := <-self.result
	p.Block.Header.SigData = result
	log.Info("To consensus successfully with signature %x.", result)
	return nil
}

func (self *BFTPolicy) Halt() {
	return
}
