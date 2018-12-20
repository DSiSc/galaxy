package consensus

import (
	"fmt"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/config"
	"github.com/DSiSc/validator/tools/account"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func mockConf(policy string) config.ConsensusConfig {
	return config.ConsensusConfig{
		PolicyName: policy,
	}
}

var mockAccount = account.Account{
	Address: types.Address{0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68,
		0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
	Extension: account.AccountExtension{
		Id:  0,
		Url: "172.0.0.1:8080",
	},
}

func Test_NewConsensus(t *testing.T) {
	asserts := assert.New(t)
	conf := mockConf("solo")
	consensus, err := NewConsensus(conf, mockAccount, nil)
	asserts.Nil(err)
	asserts.NotNil(consensus)
	asserts.Equal("solo", consensus.PolicyName())

	p := reflect.TypeOf(consensus)
	method, exist := p.MethodByName("PolicyName")
	asserts.NotNil(method)
	asserts.True(exist)

	method, exist = p.MethodByName("ToConsensus")
	asserts.NotNil(method)
	asserts.True(exist)

	conf = mockConf(common.BftPolicy)
	consensus, err = NewConsensus(conf, mockAccount, nil)
	asserts.Equal(common.BftPolicy, consensus.PolicyName())

	conf = mockConf(common.FbftPolicy)
	consensus, err = NewConsensus(conf, mockAccount, nil)
	asserts.Equal(common.FbftPolicy, consensus.PolicyName())

	conf = mockConf(common.DbftPolicy)
	consensus, err = NewConsensus(conf, mockAccount, nil)
	asserts.Equal(common.DbftPolicy, consensus.PolicyName())

	policyName := "Nil"
	conf = mockConf(policyName)
	consensus, err = NewConsensus(conf, mockAccount, nil)
	asserts.Equal(err, fmt.Errorf("unsupport consensus type %v", policyName))
}
