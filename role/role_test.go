package role

import (
	"github.com/DSiSc/galaxy/role/config"
	"github.com/DSiSc/txpool/common"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func mock_address(num int) []common.Address {
	to := make([]common.Address, num)
	for m := 0; m < num; m++ {
		for j := 0; j < common.AddressLength; j++ {
			to[m][j] = byte(m)
		}
	}
	return to
}

func mock_conf() *config.RoleConfig {
	return &config.RoleConfig{
		PolicyName: "solo",
	}
}

func Test_NewRole(t *testing.T) {
	assert := assert.New(t)
	address := mock_address(1)[0]
	assert.NotNil(address)
	conf := mock_conf()
	role, err := NewRole(nil, address, conf)
	assert.Nil(err)
	assert.NotNil(role)

	p := reflect.TypeOf(role)
	method, exist := p.MethodByName("PolicyName")
	assert.NotNil(method)
	assert.True(exist)

	method, exist = p.MethodByName("RoleAssignments")
	assert.NotNil(method)
	assert.True(exist)

	method, exist = p.MethodByName("GetRoles")
	assert.NotNil(method)
	assert.True(exist)
}
