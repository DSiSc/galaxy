package policy

import (
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/justitia/config"
	"github.com/DSiSc/txpool/log"
)

const (
	SOLO_POLICY = "solo"
)

type SoloPolicy struct {
	name string
}

func NewSoloPolicy() (*SoloPolicy, error) {
	return &SoloPolicy{name: SOLO_POLICY}, nil
}

func (self *SoloPolicy) PolicyName() string {
	return self.name
}

func (self *SoloPolicy) getMembers() types.NodeAddress {
	nodeName := config.SINGLE_NODE_NAME
	return types.NodeAddress(nodeName)
}

func (self *SoloPolicy) GetParticipates() ([]types.NodeAddress, error) {
	participates := make([]types.NodeAddress, 0, 1)
	member := self.getMembers()
	participates = append(participates, member)
	return participates, nil
}

func (self *SoloPolicy) ChangeParticipates() error {
	log.Info("Solo will not change participate.")
	return nil
}
