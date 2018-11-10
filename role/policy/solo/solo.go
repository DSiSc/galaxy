package solo

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/validator/tools/account"
)

type SoloPolicy struct {
	name         string
	assignments  map[account.Account]common.Roler
	participates participates.Participates
}

func NewSoloPolicy(participates participates.Participates) (*SoloPolicy, error) {
	soloPolicy := &SoloPolicy{
		name:         common.SOLO_POLICY,
		participates: participates,
	}
	return soloPolicy, nil
}

func (self *SoloPolicy) RoleAssignments() (map[account.Account]common.Roler, error) {
	accounts, err := self.participates.GetParticipates()
	if nil != err {
		log.Error("get participates failed with error %v.", err)
		return nil, fmt.Errorf("get participates failed")
	}
	participates := len(accounts)
	if 1 != participates {
		log.Error("solo role policy only support one participate.")
		return nil, fmt.Errorf("more than one participate")
	}
	self.assignments = make(map[account.Account]common.Roler, participates)
	self.assignments[accounts[0]] = common.Master
	return self.assignments, nil
}

func (self *SoloPolicy) GetRoles(address account.Account) (common.Roler, error) {
	if 0 == len(self.assignments) {
		log.Error("RoleAssignments must be called before.")
		return common.UnKnown, common.AssignmentNotBeExecute
	}
	if role, ok := self.assignments[address]; !ok {
		log.Error("wrong address which nobody knows in solo policy")
		return common.UnKnown, fmt.Errorf("wrong address")
	} else {
		return role, nil
	}
}

func (self *SoloPolicy) PolicyName() string {
	return self.name
}
