package consensus

import (
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/config"
	"github.com/DSiSc/galaxy/consensus/policy"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/txpool/common/log"
)

type Consensus interface {
	PolicyName() string
	ToConsensus(p *common.Proposal) (bool, error)
}

const (
	CONSENSUS_SOLO = "solo"
	// Structure must matching with defination of config/config.json
	Symbol = "consensus"
	Policy = "consensus.policy"
)

func NewConsensus(participates participates.Participates, conf config.ConsensusConfig) (Consensus, error) {
	var err error
	var consensus Consensus
	consensusPolicy := conf.PolicyName
	switch consensusPolicy {
	case CONSENSUS_SOLO:
		log.Info("Get consensus policy is solo.")
		consensus, err = policy.NewSoloPolicy(participates)
	default:
		log.Error("Now, we only support solo policy consensus.")
	}
	return consensus, err
}
