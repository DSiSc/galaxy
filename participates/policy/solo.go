package policy

import (
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/validator/tools/account"
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

func (self *SoloPolicy) getMembers() account.Account {
	address := types.Address{
		0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33,
		0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d,
	}

	return account.Account{
		Address: address,
	}
}

func (self *SoloPolicy) GetParticipates() ([]account.Account, error) {
	participates := make([]account.Account, 0, 1)
	member := self.getMembers()
	participates = append(participates, member)
	return participates, nil
}

func (self *SoloPolicy) ChangeParticipates() error {
	log.Warn("Solo will not change participate.")
	return nil
}
