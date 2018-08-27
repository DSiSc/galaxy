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
	local       types.NodeAddress
	participate participates.Participates
}

func NewSoloPolicy(p participates.Participates, localNode types.NodeAddress) (*SoloPolicy, error) {
	soloPolicy := &SoloPolicy{
		name:        SOLO_POLICY,
		local:       localNode,
		participate: p,
	}
	return soloPolicy, nil
}

func (self *SoloPolicy) RoleAssignments() (map[types.NodeAddress]common.Roler, error) {
	members, err := self.participate.GetParticipates()
	if err != nil {
		log.Error("Error to get participates.")
		return nil, fmt.Errorf("Get participates with error:%s.", err)
	}

	if len(members) != 1 {
		log.Error("Solo role policy must match solo participates policy.")
		return nil, fmt.Errorf("Participates policy not match solo role policy.")
	}

	assigments := map[types.NodeAddress]common.Roler{
		members[0]: common.Master,
	}
	return assigments, nil
}

func (self *SoloPolicy) GetRoles(address types.NodeAddress) common.Roler {
	if address != self.local {
		log.Error("Wrong address which nobody knows in solo role policy.")
		return common.UnKnown
	}
	return common.Master
}

func (self *SoloPolicy) PolicyName() string {
	return self.name
}
