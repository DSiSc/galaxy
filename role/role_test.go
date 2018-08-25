package role

import (
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/role/config"
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

func Test_NewRole(t *testing.T) {
	asserts := assert.New(t)
	address := mock_address(1)[0]
	asserts.NotNil(address)
	conf := mock_conf()
	role, err := NewRole(nil, address, conf)
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
}
