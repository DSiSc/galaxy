package consensus

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/config"
	"github.com/DSiSc/galaxy/consensus/policy/bft"
	"github.com/DSiSc/galaxy/consensus/policy/dbft"
	"github.com/DSiSc/galaxy/consensus/policy/fbft"
	"github.com/DSiSc/galaxy/consensus/policy/solo"
	"github.com/DSiSc/validator/tools/account"
)

type Consensus interface {
	PolicyName() string
	Initialization(account.Account, account.Account, []account.Account, types.EventCenter, bool)
	ToConsensus(p *common.Proposal) error
	GetConsensusResult() common.ConsensusResult
	Online()
	Start()
	Halt()
}

func NewConsensus(conf config.ConsensusConfig, blockSwitch chan<- interface{}) (Consensus, error) {
	var err error
	var consensus Consensus
	switch conf.PolicyName {
	case common.SoloPolicy:
		log.Info("Get consensus policy is solo.")
		consensus, err = solo.NewSoloPolicy(blockSwitch, conf.EnableEmptyBlock, conf.SignVerifySwitch)
	case common.BftPolicy:
		log.Info("Get consensus policy is bft.")
		consensus, err = bft.NewBFTPolicy(conf.Timeout)
	case common.FbftPolicy:
		log.Info("Get consensus policy is fbft.")
		consensus, err = fbft.NewFBFTPolicy(conf.Timeout, blockSwitch, conf.EnableEmptyBlock, conf.SignVerifySwitch)
	case common.DbftPolicy:
		log.Info("Get consensus policy is dbft.")
		consensus, err = dbft.NewDBFTPolicy(conf.Timeout)
	default:
		err = fmt.Errorf("unsupport consensus type %v", conf.PolicyName)
	}
	return consensus, err
}
