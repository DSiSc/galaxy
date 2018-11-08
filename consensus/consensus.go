package consensus

import (
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/config"
	"github.com/DSiSc/galaxy/consensus/policy/solo"
	"github.com/DSiSc/galaxy/participates"
)

type Consensus interface {
	PolicyName() string
	ToConsensus(p *common.Proposal) error
	Start()
	Halt()
}

func NewConsensus(participates participates.Participates, conf config.ConsensusConfig) (Consensus, error) {
	var err error
	var consensus Consensus
	consensusPolicy := conf.PolicyName
	switch consensusPolicy {
	case common.SOLO_POLICY:
		log.Info("Get consensus policy is solo.")
		consensus, err = solo.NewSoloPolicy(participates)
	default:
		log.Error("Now, we only support solo policy consensus.")
	}
	return consensus, err
}
