package solo

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/validator/tools/account"
)

type SoloPolicy struct {
	name         string
	participates []account.Account
}

func NewSoloPolicy(accounts []account.Account) (*SoloPolicy, error) {
	if 1 != len(accounts) {
		log.Error("solo role policy only support one participate.")
		return nil, fmt.Errorf("solo role policy only support one participate")
	}
	soloPolicy := &SoloPolicy{
		name:         common.SOLO_POLICY,
		participates: accounts,
	}
	return soloPolicy, nil
}

func (self *SoloPolicy) RoleAssignments() (map[account.Account]common.Roler, error) {
	participates := len(self.participates)
	if 1 != participates {
		log.Error("solo role policy only support one participate.")
		return nil, fmt.Errorf("more than one participate")
	}
	assignment := make(map[account.Account]common.Roler, participates)
	assignment[self.participates[0]] = common.Master
	return assignment, nil
}

func (self *SoloPolicy) GetRoles(address account.Account) common.Roler {
	if address != self.participates[0] {
		log.Error("wrong address which nobody knows in solo role policy")
		return common.UnKnown
	}
	return common.Master
}

func (self *SoloPolicy) PolicyName() string {
	return self.name
}
