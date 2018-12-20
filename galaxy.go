package galaxy

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/galaxy/common"
	"github.com/DSiSc/galaxy/consensus"
	"github.com/DSiSc/galaxy/role"
)

func NewGalaxyPlugin(conf common.GalaxyPluginConf) (*common.GalaxyPlugin, error) {
	role, err := role.NewRole(conf.RoleConf)
	if nil != err {
		log.Error("Init role failed.")
		return nil, fmt.Errorf("role init failed")
	}
	consensus, err := consensus.NewConsensus(conf.ConsensusConf, conf.Account, conf.BlockSwitch)
	if nil != err {
		log.Error("Init consensus failed.")
		return nil, fmt.Errorf("consensus init failed")
	}
	return &common.GalaxyPlugin{
		Role:      role,
		Consensus: consensus,
	}, err
}
