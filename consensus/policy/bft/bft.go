package bft

import (
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

func NewBFTPolicy(participate participates.Participates, account account.Account, master account.Account) (*BFTPolicy, error) {
	members, _ := participate.GetParticipates()
	policy := &BFTPolicy{
		name:         common.BFT_POLICY,
		account:      account,
		participates: participate,
		bftCore:      NewBFTCore(account.Extension.Id, master.Extension.Id, members),
	}
	return policy, nil
}

func (self *BFTPolicy) PolicyName() string {
	return self.name
}

func (self *BFTPolicy) Start() {
	log.Info("start bft policy service.")
	self.bftCore.Start(self.account)
}
