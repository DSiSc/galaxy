package participates

import (
	"fmt"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/participates/config"
	"github.com/DSiSc/galaxy/participates/policy"
	"github.com/DSiSc/txpool/log"
)

type Participates interface {
	PolicyName() string
	GetParticipates() ([]types.Address, error)
	ChangeParticipates() error
}

const (
	PARTICIPATES_SOLO = "solo"
)

func NewParticipates(conf config.ParticipateConfig) (Participates, error) {
	var err error
	var participates Participates
	participatesPolicy := conf.PolicyName
	switch participatesPolicy {
	case PARTICIPATES_SOLO:
		log.Info("Get participates policy is solo.")
		participates, err = policy.NewSoloPolicy()
	default:
		log.Error("Now, we only support solo policy participates.")
		err = fmt.Errorf("Not support type.")
	}
	return participates, err
}
