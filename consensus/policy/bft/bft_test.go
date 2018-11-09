package bft

import (
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/participates/config"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/validator/tools/account"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func mock_conf(policy string) config.ParticipateConfig {
	return config.ParticipateConfig{
		PolicyName: policy,
		Delegates:  4,
	}
}

func TestNewBFTPolicy(t *testing.T) {
	conf := mock_conf("dpos")
	participate, err := participates.NewParticipates(conf)
	assert.Nil(t, err)
	bft, err := NewBFTPolicy(participate, mockAccounts[0], mockAccounts[1])
	assert.NotNil(t, bft)
	assert.Nil(t, err)
	assert.Equal(t, common.BFT_POLICY, bft.name)
	assert.NotNil(t, bft.bftCore)
	assert.Equal(t, uint8((conf.Delegates-1)/3), bft.bftCore.tolerance)
	assert.Equal(t, mockAccounts[0].Extension.Id, bft.bftCore.id)
	assert.Equal(t, mockAccounts[1].Extension.Id, bft.bftCore.master)
}

func TestBFTPolicy_PolicyName(t *testing.T) {
	conf := mock_conf("dpos")
	participate, err := participates.NewParticipates(conf)
	assert.Nil(t, err)
	bft, _ := NewBFTPolicy(participate, mockAccounts[0], mockAccounts[1])
	assert.Equal(t, common.BFT_POLICY, bft.name)
	assert.Equal(t, bft.name, bft.PolicyName())
}

func TestBFTPolicy_Start(t *testing.T) {
	conf := mock_conf("dpos")
	participate, err := participates.NewParticipates(conf)
	assert.Nil(t, err)
	bft, _ := NewBFTPolicy(participate, mockAccounts[0], mockAccounts[1])
	var b *bftCore
	monkey.PatchInstanceMethod(reflect.TypeOf(b), "Start", func(*bftCore, account.Account) {
		log.Info("pass it.")
		return
	})
	bft.Start()
}
