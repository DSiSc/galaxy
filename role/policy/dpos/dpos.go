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

func NewDPOSPolicy() (*DPOSPolicy, error) {
	policy := &DPOSPolicy{
		name: common.DposPolicy,
	}
	return policy, nil
}

func (instance *DPOSPolicy) RoleAssignments(participates []account.Account) (map[account.Account]common.Roler, account.Account, error) {
	block, ok := blockchain.NewLatestStateBlockChain()
	if nil != ok {
		log.Error("Get NewLatestStateBlockChain failed.")
		return nil, account.Account{}, fmt.Errorf("get NewLatestStateBlockChain failed")
	}
	instance.participates = participates
	delegates := len(instance.participates)
	instance.assignments = make(map[account.Account]common.Roler, delegates)
	currentBlockHeight := block.GetCurrentBlock().Header.Height
	masterId := (currentBlockHeight + 1) % uint64(delegates)
	// masterIndex := currentBlockHeight % uint64(delegates)
	var master account.Account
	for _, delegate := range instance.participates {
		if delegate.Extension.Id == masterId {
			instance.assignments[delegate] = common.Master
			master = delegate
		} else {
			instance.assignments[delegate] = common.Slave
		}
	}
	return instance.assignments, master, nil
}

func (instance *DPOSPolicy) GetRoles(account account.Account) (common.Roler, error) {
	if 0 == len(instance.assignments) {
		log.Error("role assignment has not been executed.")
		return common.UnKnown, common.AssignmentNotBeExecute
	}
	if role, ok := instance.assignments[account]; !ok {
		log.Error("account %x is not a delegate, please confirm.", account)
		// TODO: verify normal node or unknown
		return common.UnKnown, fmt.Errorf("accont not a delegate")
	} else {
		log.Info("account %x role is %v.", account, role)
		return role, nil
	}
}

func (instance *DPOSPolicy) PolicyName() string {
	return instance.name
}

func (instance *DPOSPolicy) AppointRole(master account.Account) error {
	if _, ok := instance.assignments[master]; !ok {
		log.Error("account %x has not assign role, please confirm.", master)
		return fmt.Errorf("appoint account is not a delegate")
	}
	var preMaster account.Account
	var exist bool = false
	for delegate, role := range instance.assignments {
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
	instance.assignments[master] = common.Master
	instance.assignments[preMaster] = common.Slave
	return nil
}

func (instance *DPOSPolicy) ChangeRoleAssignment(assignments map[account.Account]common.Roler, master uint64) {
	for account, _ := range assignments {
		if account.Extension.Id == master {
			assignments[account] = common.Master
			continue
		}
		assignments[account] = common.Slave
	}
	instance.assignments = assignments
}
