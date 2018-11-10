package dpos

import (
	"fmt"
	"github.com/DSiSc/blockchain"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/validator/tools/account"
)

type DPOSPolicy struct {
	name         string
	participates participates.Participates
	assignments  map[account.Account]common.Roler
}

func NewDPOSPolicy(participates participates.Participates) (*DPOSPolicy, error) {
	policy := &DPOSPolicy{
		name:         common.DPOS_POLICY,
		participates: participates,
	}
	return policy, nil
}

func (self *DPOSPolicy) RoleAssignments() (map[account.Account]common.Roler, error) {
	accounts, err := self.participates.GetParticipates()
	if nil != err {
		log.Error("get participates failed with error %v.", err)
		return nil, fmt.Errorf("get participates failed")
	}
	// TODO: simply we decide master by block height, while it will support appoint external
	block, ok := blockchain.NewLatestStateBlockChain()
	if nil != ok {
		log.Error("Get NewLatestStateBlockChain failed.")
		return nil, fmt.Errorf("get NewLatestStateBlockChain failed")
	}
	delegates := len(accounts)
	self.assignments = make(map[account.Account]common.Roler, delegates)
	currentBlockHeight := block.GetCurrentBlock().Header.Height
	masterIndex := (currentBlockHeight + 1) % uint64(delegates)
	for index, delegate := range accounts {
		if index == int(masterIndex) {
			self.assignments[delegate] = common.Master
		} else {
			self.assignments[delegate] = common.Slave
		}
	}
	return self.assignments, nil
}

func (self *DPOSPolicy) GetRoles(account account.Account) (common.Roler, error) {
	if 0 == len(self.assignments) {
		log.Error("role assignment has not been executed.")
		return common.UnKnown, common.AssignmentNotBeExecute
	}
	if role, ok := self.assignments[account]; !ok {
		log.Error("account %x is not a delegate, please confirm.", account)
		// TODO: verify normal node or unknown
		return common.UnKnown, fmt.Errorf("accont not a delegate")
	} else {
		log.Info("account %x role is %v.", account, role)
		return role, nil
	}
}

func (self *DPOSPolicy) PolicyName() string {
	return self.name
}

func (self *DPOSPolicy) AppointRole(master account.Account) error {
	if _, ok := self.assignments[master]; !ok {
		log.Error("account %x has not assign role, please confirm.", master)
		return fmt.Errorf("appoint account is not a delegate")
	}
	var preMaster account.Account
	var exist bool = false
	for delegate, role := range self.assignments {
		if common.Master == role {
			preMaster = delegate
			exist = true
			break
		}
	}
	if !exist {
		log.Error("no master in delegates, please confirm.")
		return fmt.Errorf("no master exist in current delegates")
	}
	self.assignments[master] = common.Master
	self.assignments[preMaster] = common.Slave
	return nil
}
