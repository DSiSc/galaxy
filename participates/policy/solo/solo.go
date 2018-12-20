package solo

import (
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/participates/common"
	"github.com/DSiSc/validator/tools/account"
)

type SoloPolicy struct {
	name string
}

func NewSoloPolicy() (*SoloPolicy, error) {
	return &SoloPolicy{name: common.SoloPolicy}, nil
}

func (instance *SoloPolicy) PolicyName() string {
	return instance.name
}

func (instance *SoloPolicy) GetParticipates() ([]account.Account, error) {
	participates := make([]account.Account, 0, 1)
	participate := account.Account{
		Address: types.Address{
			0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33,
			0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d,
		},
		Extension: account.AccountExtension{
			Id:  0,
			Url: "127.0.0.1:8080",
		},
	}
	participates = append(participates, participate)
	return participates, nil
}
