package solo

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/validator/tools/account"
)

type SoloPolicy struct {
	name        string
	local       account.Account
	participate participates.Participates
}

func NewSoloPolicy(p participates.Participates, localNode account.Account) (*SoloPolicy, error) {
	soloPolicy := &SoloPolicy{
		name:        common.SOLO_POLICY,
		local:       localNode,
		participate: p,
	}
	return soloPolicy, nil
}

func (self *SoloPolicy) RoleAssignments() (map[account.Account]common.Roler, error) {
	assignment := make(map[account.Account]common.Roler, 1)
	members, err := self.participate.GetParticipates()
	if err != nil {
		log.Error("Error to get participates.")
		return nil, fmt.Errorf("get participates with error:%s", err)
	}

	if len(members) != 1 {
		log.Error("Solo role policy must match solo participates policy.")
		return nil, fmt.Errorf("participates policy not match solo role policy")
	}

	if members[0].Address != self.local.Address {
		log.Error("Solo role policy only support local account.")
		return nil, fmt.Errorf("solo role policy only support local account")
	}

	assignment[self.local] = common.Master
	return assignment, nil
}

func (self *SoloPolicy) GetRoles(address account.Account) common.Roler {
	if address != self.local {
		log.Error("wrong address which nobody knows in solo role policy")
		return common.UnKnown
	}
	return common.Master
}

func (self *SoloPolicy) PolicyName() string {
	return self.name
}
