package role

import (
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/role/config"
	"github.com/DSiSc/validator/tools/account"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func mock_address(num int) []types.Address {
	to := make([]types.Address, num)
	for m := 0; m < num; m++ {
		for j := 0; j < types.AddressLength; j++ {
			to[m][j] = byte(m)
		}
	}
	return to
}

func mock_conf() config.RoleConfig {
	return config.RoleConfig{
		PolicyName: "solo",
	}
}

var MockAccount = account.Account{
	Address: types.Address{
		0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33,
		0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d,
	},
}

func Test_NewRole(t *testing.T) {
	asserts := assert.New(t)
	conf := mock_conf()
	role, err := NewRole(nil, MockAccount, conf)
	asserts.Nil(err)
	asserts.NotNil(role)

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

	fakeConf := config.RoleConfig{
		PolicyName: "unknown",
	}
	role, err = NewRole(nil, MockAccount, fakeConf)
	asserts.NotNil(err)
	asserts.Nil(role)
}
