package bft

import (
	"fmt"
	"github.com/DSiSc/galaxy/consensus/common"
	commonr "github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/validator/tools/account"
	"github.com/ontio/ontology/common/log"
)

type BFTPolicy struct {
	name string
	// local account
	account account.Account
	bftCore *bftCore
}

func NewBFTPolicy(account account.Account) (*BFTPolicy, error) {
	policy := &BFTPolicy{
		name:    common.BFT_POLICY,
		account: account,
		bftCore: NewBFTCore(account.Extension.Id),
	}
	return policy, nil
}

func (self *BFTPolicy) Prepare(role map[account.Account]commonr.Roler, peers []account.Account) error {
	if len(role) != len(peers) {
		log.Error("bft core has not been initial, please confirm.")
		return fmt.Errorf("role and peers not in consistent")
	}

	var masterExist bool = false
	for delegate, role := range role {
		if commonr.Master == role {
			self.bftCore.master = delegate.Extension.Id
			masterExist = true
		}
	}
	if !masterExist {
		log.Errorf("no master exist in delegates")
		return fmt.Errorf("no master")
	}

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

func (self *BFTPolicy) ToConsensus(p *common.Proposal) error {
	// TODO: send request
	return nil
}

func (self *BFTPolicy) Halt() {
	return
}

func (self *BFTPolicy) final() ([][]byte, []account.Account, error) {
	result := <-self.bftCore.result
	// check result
	signData := result.signatures
	signMap := result.signMap
	if len(signData) != len(signMap) {
		log.Error("length of signData[%d] and signMap[%d] does not match.", len(signData), len(signMap))
		return nil, nil, fmt.Errorf("result not in coincidence")
	}
	var reallySignature = make([][]byte, 0)
	var suspiciousAccount = make([]account.Account, 0)
	for account, sign := range signMap {
		if signDataVerify(account, sign) {
			reallySignature = append(reallySignature, sign)
			continue
		}
		suspiciousAccount = append(suspiciousAccount, account)
		log.Error("signature %x by account %x is invalid", sign, account)
	}
	return reallySignature, suspiciousAccount, nil
}

func signDataVerify(account account.Account, sign []byte) bool {
	return true
}
