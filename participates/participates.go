package participates

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/galaxy/participates/config"
	"github.com/DSiSc/galaxy/participates/policy"
	"github.com/DSiSc/validator/tools/account"
)

type Participates interface {
	PolicyName() string
	GetParticipates() ([]account.Account, error)
	ChangeParticipates() error
}

func NewParticipates(conf config.ParticipateConfig) (Participates, error) {
	var err error
	var participates Participates
	participatesPolicy := conf.PolicyName
	switch participatesPolicy {
	case policy.SOLO_POLICY:
		log.Info("Get participates policy is solo.")
		participates, err = policy.NewSoloPolicy()
	default:
		log.Error("Now, we only support solo policy participates.")
		err = fmt.Errorf("Not support type.")
	}
	return participates, err
}
