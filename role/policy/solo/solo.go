package solo

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/validator/tools/account"
)

type SoloPolicy struct {
	name         string
	assignments  map[account.Account]common.Roler
	participates []account.Account
}

func NewSoloPolicy() (*SoloPolicy, error) {
	soloPolicy := &SoloPolicy{
		name: common.SoloPolicy,
	}
	return soloPolicy, nil
}

func (instance *SoloPolicy) RoleAssignments(accounts []account.Account) (map[account.Account]common.Roler, account.Account, error) {
	participates := len(accounts)
	if 1 != participates {
		log.Error("solo role policy only support one participate.")
		return nil, account.Account{}, fmt.Errorf("more than one participate")
	}
	instance.participates = accounts
	instance.assignments = make(map[account.Account]common.Roler, participates)
	instance.assignments[instance.participates[0]] = common.Master
	return instance.assignments, instance.participates[0], nil
}

func (instance *SoloPolicy) GetRoles(address account.Account) (common.Roler, error) {
	if 0 == len(instance.assignments) {
		log.Error("RoleAssignments must be called before.")
		return common.UnKnown, common.AssignmentNotBeExecute
	}
	if role, ok := instance.assignments[address]; !ok {
		log.Error("wrong address which nobody knows in solo policy")
		return common.UnKnown, fmt.Errorf("wrong address")
	} else {
		return role, nil
	}
}

func (instance *SoloPolicy) PolicyName() string {
	return instance.name
}

func (instance *SoloPolicy) ChangeRoleAssignment(assignments map[account.Account]common.Roler, master uint64) {
	for account, _ := range assignments {
		if account.Extension.Id == master {
			assignments[account] = common.Master
			continue
		}
		assignments[account] = common.Slave
	}
	instance.assignments = assignments
}
