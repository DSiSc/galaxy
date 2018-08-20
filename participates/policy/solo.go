package policy

import (
	"github.com/DSiSc/txpool/common"
	"github.com/DSiSc/txpool/common/log"
)

const (
	POLICY_NAME = "solo"
)

type SoloPolicy struct {
	name string
}

func NewSoloPolicy() (*SoloPolicy, error) {
	return &SoloPolicy{name: POLICY_NAME}, nil
}

func (self *SoloPolicy) PolicyName() string {
	return self.name
}

func (self *SoloPolicy) GetParticipates() []common.Address {
	participates := make([]common.Address, 0, 0)
	log.Info("Solo will return nil when getting participates.")
	return participates
}

func (self *SoloPolicy) ChangeParticipates() error {
	log.Info("Solo will not change participate.")
	return nil
}
