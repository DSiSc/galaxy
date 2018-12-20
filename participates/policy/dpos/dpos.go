package dpos

import (
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/participates/common"
	"github.com/DSiSc/validator/tools/account"
)

type DPOSPolicy struct {
	name string
	// number of delegates
	members      uint64
	participates []account.Account
}

func NewDPOSPolicy(number uint64) (*DPOSPolicy, error) {
	return &DPOSPolicy{
		name:    common.DposPolicy,
		members: number,
	}, nil
}

func (instance *DPOSPolicy) PolicyName() string {
	return instance.name
}

// Get the top ranking of count from voting result.
func (instance *DPOSPolicy) getDelegatesByCount(count uint64) ([]account.Account, error) {
	// TODO: Get accounts by voting result
	account0 := account.Account{
		Address: types.Address{0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  0,
			Url: "192.168.176.145:8080",
		},
	}

	account1 := account.Account{
		Address: types.Address{0x34, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  1,
			Url: "192.168.176.146:8080",
		},
	}

	account2 := account.Account{
		Address: types.Address{0x35, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  2,
			Url: "192.168.176.147:8080",
		},
	}

	account3 := account.Account{
		Address: types.Address{0x36, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  3,
			Url: "192.168.176.148:8080",
		},
	}

	accounts := []account.Account{account0, account1, account2, account3}
	return accounts, nil
}

func (instance *DPOSPolicy) getDelegates() ([]account.Account, error) {
	delegates, err := instance.getDelegatesByCount(instance.members)
	if nil != err {
		log.Error("get delegates failed with err %s.", err)
		return nil, err
	}
	return delegates, nil
}

func (instance *DPOSPolicy) GetParticipates() ([]account.Account, error) {
	participates, err := instance.getDelegates()
	if nil != err {
		log.Error("Get delegates failed with error %v.", err)
	} else {
		instance.participates = participates
	}
	return participates, err
}
