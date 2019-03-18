package dpos

import (
	"github.com/DSiSc/contractsManage/contracts"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/participates/common"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/validator/tools/account"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

var delegates uint64 = 4

var accounts = []account.Account{
	account.Account{
		Address: types.Address{0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  0,
			Url: "192.168.176.145:8080",
		},
	},
	account.Account{
		Address: types.Address{0x34, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  1,
			Url: "192.168.176.146:8080",
		},
	},
	account.Account{
		Address: types.Address{0x35, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  2,
			Url: "192.168.176.147:8080",
		},
	},
	account.Account{
		Address: types.Address{0x36, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  3,
			Url: "192.168.176.148:8080",
		},
	},
}

var mockNodes = []contracts.NodeInfo{
	contracts.NodeInfo{
		Address: types.Address{0x35, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Id:      0,
		Url:     "192.168.176.145:8080",
	},
	contracts.NodeInfo{
		Address: types.Address{0x36, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Id:      1,
		Url:     "192.168.176.146:8080",
	},
	contracts.NodeInfo{
		Address: types.Address{0x37, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Id:      2,
		Url:     "192.168.176.147:8080",
	},
	contracts.NodeInfo{
		Address: types.Address{0x38, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Id:      3,
		Url:     "192.168.176.148:8080",
	},
}

func TestDPOSPolicy_PolicyName(t *testing.T) {
	asserts := assert.New(t)
	monkey.Patch(contracts.NewVotingContract, func() contracts.Voting {
		return &contracts.VotingContract{}
	})
	dpos := NewDPOSPolicy()
	asserts.NotNil(dpos)
	asserts.Equal(common.DposPolicy, dpos.PolicyName())
	asserts.Equal(uint64(0), dpos.members)
	asserts.Equal(uint64(0), uint64(len(dpos.participates)))
}

func TestNewDPOSPolicy(t *testing.T) {
	asserts := assert.New(t)
	monkey.Patch(contracts.NewVotingContract, func() contracts.Voting {
		return &contracts.VotingContract{}
	})
	dpos := NewDPOSPolicy()
	asserts.NotNil(dpos)
	asserts.Equal(dpos.name, common.DposPolicy)
}

func TestDPOSPolicy_GetParticipates(t *testing.T) {
	asserts := assert.New(t)
	var contract *contracts.VotingContract
	monkey.Patch(contracts.NewVotingContract, func() contracts.Voting {
		return contract
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(contract), "NodeNumber", func(*contracts.VotingContract) uint64 {
		return uint64(4)
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(contract), "GetNodeList", func(*contracts.VotingContract, uint64) ([]contracts.NodeInfo, error) {
		return mockNodes, nil
	})
	dpos := NewDPOSPolicy()
	asserts.NotNil(dpos)
	participates, err := dpos.GetParticipates()
	asserts.Nil(err)
	asserts.Equal(delegates, dpos.members)
	asserts.Equal(dpos.members, uint64(len(dpos.participates)))
	asserts.Equal(participates, dpos.participates)
}
