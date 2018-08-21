package role

import (
	txpoolc "github.com/DSiSc/txpool/common"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func mock_address(num int) []txpoolc.Address {
	to := make([]txpoolc.Address, num)
	for m := 0; m < num; m++ {
		for j := 0; j < txpoolc.AddressLength; j++ {
			to[m][j] = byte(m)
		}
	}
	return to
}

func Test_NewRolePolicy(t *testing.T) {
	assert := assert.New(t)
	address := mock_address(1)[0]
	assert.NotNil(address)
	role, err := NewRolePolicy(nil, address)
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
