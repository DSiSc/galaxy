package consensus

import (
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/config"
	"github.com/DSiSc/galaxy/consensus/policy/bft"
	"github.com/DSiSc/galaxy/consensus/policy/bft/dbft"
	"github.com/DSiSc/galaxy/consensus/policy/fbft"
	"github.com/DSiSc/galaxy/consensus/policy/solo"
	"github.com/DSiSc/galaxy/participates"
	commonr "github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/validator/tools/account"
)

type Consensus interface {
	PolicyName() string
	Initialization(map[account.Account]commonr.Roler, []account.Account, types.EventCenter) error
	ToConsensus(p *common.Proposal) error
	GetConsensusResult() common.ConsensusResult
	Start()
	Halt()
}

func NewConsensus(participates participates.Participates, conf config.ConsensusConfig, account account.Account, blockSwitch chan<- interface{}) (Consensus, error) {
	var err error
	var consensus Consensus
	switch conf.PolicyName {
	case common.SOLO_POLICY:
		log.Info("Get consensus policy is solo.")
		consensus, err = solo.NewSoloPolicy(account, blockSwitch)
	case common.BFT_POLICY:
		log.Info("Get consensus policy is bft.")
		consensus, err = bft.NewBFTPolicy(account, conf.Timeout)
	case common.FBFT_POLICY:
		log.Info("Get consensus policy is fbft.")
		consensus, err = fbft.NewFBFTPolicy(account, conf.Timeout, blockSwitch)
	case common.DBFT_POLICY:
		log.Info("Get consensus policy is dbft.")
		consensus, err = dbft.NewDBFTPolicy(account, conf.Timeout)
	default:
		log.Error("Now, we only support solo policy consensus.")
	}
	return consensus, err
}
