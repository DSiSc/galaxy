package consensus

import (
	"github.com/DSiSc/galaxy/consensus/config"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func mock_conf() *config.ConsensusPolicy {
	return &config.ConsensusPolicy{
		PolicyName: "solo",
	}
}

func Test_NewConsensusPolicy(t *testing.T) {
	assert := assert.New(t)
	conf := mock_conf()
	consensus, err := NewConsensusPolicy(nil, conf)
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
