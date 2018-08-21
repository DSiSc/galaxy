package consensus

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_NewConsensusPolicy(t *testing.T) {
	assert := assert.New(t)

	consensus, err := NewConsensusPolicy(nil)
	assert.Nil(err)
	assert.NotNil(consensus)

	p := reflect.TypeOf(consensus)
	method, exist := p.MethodByName("PolicyName")
	assert.NotNil(method)
	assert.True(exist)

	method, exist = p.MethodByName("ToConsensus")
	assert.NotNil(method)
	assert.True(exist)
}
