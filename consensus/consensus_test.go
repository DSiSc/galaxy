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
	asserts := assert.New(t)
	conf := mock_conf()
	consensus, err := NewConsensus(nil, conf)
	asserts.Nil(err)
	asserts.NotNil(consensus)

	p := reflect.TypeOf(consensus)
	method, exist := p.MethodByName("PolicyName")
	asserts.NotNil(method)
	asserts.True(exist)

	method, exist = p.MethodByName("ToConsensus")
	asserts.NotNil(method)
	asserts.True(exist)
}
