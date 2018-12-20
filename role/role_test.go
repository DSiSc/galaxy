package role

import (
	"fmt"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/galaxy/role/config"
	"github.com/DSiSc/validator/tools/account"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

var mockAccounts = []account.Account{
	account.Account{
		Address: types.Address{0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68,
			0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  0,
			Url: "172.0.0.1:8080",
		},
	},
	account.Account{
		Address: types.Address{0x34, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68,
			0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  1,
			Url: "172.0.0.1:8081"},
	},
}

func mock_solo_conf() config.RoleConfig {
	return config.RoleConfig{
		PolicyName: "solo",
	}
}

func mock_dpos_conf() config.RoleConfig {
	return config.RoleConfig{
		PolicyName: "dpos",
	}
}

func Test_NewRole(t *testing.T) {
	asserts := assert.New(t)
	conf := mock_solo_conf()
	role, err := NewRole(conf)
	asserts.Nil(err)
	asserts.NotNil(role)
	asserts.Equal(common.SoloPolicy, role.PolicyName())

	p := reflect.TypeOf(role)
	method, exist := p.MethodByName("PolicyName")
	asserts.NotNil(method)
	asserts.True(exist)

	method, exist = p.MethodByName("RoleAssignments")
	asserts.NotNil(method)
	asserts.True(exist)

	method, exist = p.MethodByName("GetRoles")
	asserts.NotNil(method)
	asserts.True(exist)

	role, err = NewRole(mock_dpos_conf())
	asserts.Nil(err)
	asserts.NotNil(role)
	asserts.Equal(common.DposPolicy, role.PolicyName())

	fakeConf := config.RoleConfig{
		PolicyName: "unknown",
	}
	role, err = NewRole(fakeConf)
	asserts.NotNil(err)
	asserts.Equal(fmt.Errorf("unkonwn policy type"), err)
	asserts.Nil(role)
}

func changeAssign(assignments map[account.Account]common.Roler, master uint64) {
	for account, _ := range assignments {
		if account.Extension.Id == master {
			assignments[account] = common.Master
			continue
		}
		assignments[account] = common.Slave
	}
}

func TestNewRole(t *testing.T) {
	var assignment map[account.Account]common.Roler
	assignment = make(map[account.Account]common.Roler)
	assignment[mockAccounts[0]] = common.Master
	assignment[mockAccounts[1]] = common.Slave

	changeAssign(assignment, 1)
	assert.Equal(t, common.Slave, assignment[mockAccounts[0]])
	assert.Equal(t, common.Master, assignment[mockAccounts[1]])
}
