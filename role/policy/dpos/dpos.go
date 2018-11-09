package dpos

import (
	"fmt"
	"github.com/DSiSc/blockchain"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/validator/tools/account"
)

type DPOSPolicy struct {
	name         string
	participates []account.Account
	assignments  map[account.Account]common.Roler
}

func NewDPOSPolicy(accounts []account.Account) (*DPOSPolicy, error) {
	policy := &DPOSPolicy{
		name:         common.DPOS_POLICY,
		participates: accounts,
	}
	return policy, nil
}

func (self *DPOSPolicy) RoleAssignments() (map[account.Account]common.Roler, error) {
	// TODO: simply we decide master by block height, while it will support appoint external
	block, ok := blockchain.NewLatestStateBlockChain()
	if nil != ok {
		log.Error("Get NewLatestStateBlockChain failed.")
		return nil, fmt.Errorf("get NewLatestStateBlockChain failed")
	}

	numberOfDelegates := len(self.participates)
	self.assignments = make(map[account.Account]common.Roler, numberOfDelegates)

	currentBlockHeight := block.GetCurrentBlock().Header.Height
	masterIndex := (currentBlockHeight + 1) % uint64(numberOfDelegates)

	for index, delegate := range self.participates {
		if index == int(masterIndex) {
			self.assignments[delegate] = common.Master
		} else {
			self.assignments[delegate] = common.Slave
		}
	}
	return self.assignments, nil
}

func (self *DPOSPolicy) GetRoles(address account.Account) common.Roler {
	if role, ok := self.assignments[address]; !ok {
		log.Error("account %x is not a delegate, please confirm.", address)
		// TODO: verify normal node or unknown
		return common.UnKnown
	} else {
		log.Info("account %x role is %v.", address, role)
		return role
	}
}

func (self *DPOSPolicy) PolicyName() string {
	return self.name
}

func (self *DPOSPolicy) AppointRole(master int) {
	if master > len(self.participates) {
		log.Error("master index %d exceed delegates span %d.", master, len(self.participates))
		return
	}

	masterAccount := self.participates[master]
	if _, ok := self.assignments[masterAccount]; !ok {
		log.Error("account %x has not assign role, please confirm.", masterAccount.Address)
		return
	}

	for delegate, role := range self.assignments {
		if delegate == masterAccount {
			if common.Master == role {
				log.Warn("Account %x has been master already.", masterAccount.Address)
				return
			} else {
				self.assignments[masterAccount] = common.Master
			}
		} else {
			self.assignments[delegate] = common.Slave
		}
	}
}
