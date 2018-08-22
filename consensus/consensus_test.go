package consensus

import (
	"github.com/DSiSc/galaxy/consensus/config"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func mock_conf() config.ConsensusConfig {
	return config.ConsensusConfig{
		PolicyName: "solo",
	}
}

func Test_NewConsensus(t *testing.T) {
	assert := assert.New(t)
	conf := mock_conf()
	consensus, err := NewConsensus(nil, conf)
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
