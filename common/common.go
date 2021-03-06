package common

import (
	"github.com/DSiSc/galaxy/consensus"
	consensusConfig "github.com/DSiSc/galaxy/consensus/config"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/participates/config"
	"github.com/DSiSc/galaxy/role"
	roleConfig "github.com/DSiSc/galaxy/role/config"
)

type GalaxyPlugin struct {
	Participates participates.Participates
	Role         role.Role
	Consensus    consensus.Consensus
}

type GalaxyPluginConf struct {
	BlockSwitch     chan<- interface{}
	ParticipateConf config.ParticipateConfig
	RoleConf        roleConfig.RoleConfig
	ConsensusConf   consensusConfig.ConsensusConfig
}
