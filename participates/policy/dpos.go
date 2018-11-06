package policy

import (
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/validator/tools/account"
)

const (
	DPOS_POLICY = "dpos"
)

type DPOSPolicy struct {
	name   string
	number uint64
}

func NewDPOSPolicy() (*DPOSPolicy, error) {
	return &DPOSPolicy{name: DPOS_POLICY}, nil
}

func (self *DPOSPolicy) PolicyName() string {
	return self.name
}

func (self *DPOSPolicy) getMembers() account.Account {
	address := types.Address{
		0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33,
		0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d,
	}

	return account.Account{
		Address: address,
	}
}

func (self *DPOSPolicy) getDelegates() ([]account.Account, error) {
	// TODO: Get Delegate Members by config or vote result
	account_0 := account.Account{
		Address: types.Address{0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  0,
			Url: "172.0.0.1:8080",
		},
	}

	account_1 := account.Account{
		Address: types.Address{0x34, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  1,
			Url: "172.0.0.1:8081",
		},
	}

	account_2 := account.Account{
		Address: types.Address{0x35, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  2,
			Url: "172.0.0.1:8082",
		},
	}

	account_3 := account.Account{
		Address: types.Address{0x36, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  3,
			Url: "172.0.0.1:8083",
		},
	}

	accounts := []account.Account{account_0, account_1, account_2, account_3}
	return accounts, nil
}

func (self *DPOSPolicy) GetParticipates() ([]account.Account, error) {
	participate, err := self.getDelegates()
	if nil != err {
		log.Error("Get delegates failed with error %v.", err)
	}
	return participate, err
}

func (self *DPOSPolicy) ChangeParticipates() error {
	// TODO: Maybe not used anymore, we can change participates by getDelegates(), which will return the latest delegates.
	return nil
}
