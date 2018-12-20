package galaxy

import (
	"fmt"
	"github.com/DSiSc/galaxy/common"
	"github.com/DSiSc/galaxy/consensus"
	consensusConfig "github.com/DSiSc/galaxy/consensus/config"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/participates/config"
	"github.com/DSiSc/galaxy/role"
	roleConfig "github.com/DSiSc/galaxy/role/config"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/validator/tools/account"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewGalaxyPlugin(t *testing.T) {
	monkey.Patch(participates.NewParticipates, func(config.ParticipateConfig) (participates.Participates, error) {
		return nil, fmt.Errorf("error of participate")
	})
	conf := common.GalaxyPluginConf{
		BlockSwitch: make(chan<- interface{}),
	}
	plugin, err := NewGalaxyPlugin(conf)
	assert.Nil(t, plugin)
	assert.Equal(t, err, fmt.Errorf("participates init failed"))

	monkey.Patch(participates.NewParticipates, func(config.ParticipateConfig) (participates.Participates, error) {
		return nil, nil
	})

	monkey.Patch(role.NewRole, func(roleConfig.RoleConfig) (role.Role, error) {
		return nil, fmt.Errorf("role init failed")
	})
	plugin, err = NewGalaxyPlugin(conf)
	assert.Nil(t, plugin)
	assert.Equal(t, err, fmt.Errorf("role init failed"))

	monkey.Patch(role.NewRole, func(roleConfig.RoleConfig) (role.Role, error) {
		return nil, nil
	})
	monkey.Patch(consensus.NewConsensus, func(consensusConfig.ConsensusConfig, account.Account, chan<- interface{}) (consensus.Consensus, error) {
		return nil, fmt.Errorf("consensus init failed")
	})
	plugin, err = NewGalaxyPlugin(conf)
	assert.Nil(t, plugin)
	assert.Equal(t, err, fmt.Errorf("consensus init failed"))

	monkey.Patch(consensus.NewConsensus, func(consensusConfig.ConsensusConfig, account.Account, chan<- interface{}) (consensus.Consensus, error) {
		return nil, nil
	})
	plugin, err = NewGalaxyPlugin(conf)
	assert.Nil(t, err)

	monkey.UnpatchAll()
}
