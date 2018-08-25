package policy

import (
	"fmt"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/txpool/log"
)

const (
	SOLO_POLICY = "solo"
)

type SoloPolicy struct {
	name        string
	local       types.Address
	participate participates.Participates
}

func NewSoloPolicy(p participates.Participates, address types.Address) (*SoloPolicy, error) {
	soloPolicy := &SoloPolicy{
		name:        SOLO_POLICY,
		local:       address,
		participate: p,
	}
	return soloPolicy, nil
}

func (self *SoloPolicy) RoleAssignments() (map[types.Address]common.Roler, error) {
	members, err := self.participate.GetParticipates()
	if err != nil {
		log.Error("Error to get participates.")
		return nil, fmt.Errorf("Get participates with error:%s.", err)
	}

	if len(members) != 0 {
		log.Error("Solo role policy must match solo participates policy.")
		return nil, fmt.Errorf("Participates policy not match solo role policy.")
	}
	return nil, nil
}

func (self *SoloPolicy) GetRoles(address types.Address) common.Roler {
	if address != self.local {
		log.Error("Wrong address which nobody knows in solo role policy.")
		return common.Unnormal
	}
	return common.Master
}

func (self *SoloPolicy) PolicyName() string {
	return self.name
}
