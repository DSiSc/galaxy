package bft

import (
	"fmt"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/validator/tools/account"
	"github.com/ontio/ontology/common/log"
)

type BFTPolicy struct {
	name         string
	account      account.Account
	participates participates.Participates
	bftCore      *bftCore
}

func NewBFTPolicy(participate participates.Participates, account account.Account) (*BFTPolicy, error) {
	policy := &BFTPolicy{
		name:         common.BFT_POLICY,
		account:      account,
		participates: participate,
		bftCore:      NewBFTCore(account.Extension.Id),
	}
	return policy, nil
}

func (self *BFTPolicy) Prepare(master account.Account, peers []account.Account) {
	// init bftCore
	if nil == self.bftCore {
		log.Error("bft core has not been initial, please confirm.")
		panic("bftCore not init")
	}
	self.bftCore.master = master.Extension.Id
	self.bftCore.peers = peers
	self.bftCore.tolerance = uint8((len(peers) - 1) / 3)
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
