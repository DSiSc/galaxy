package participates

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/galaxy/participates/common"
	"github.com/DSiSc/galaxy/participates/config"
	"github.com/DSiSc/galaxy/participates/policy/dpos"
	"github.com/DSiSc/galaxy/participates/policy/solo"
	"github.com/DSiSc/validator/tools/account"
)

type Participates interface {
	PolicyName() string
	GetParticipates() ([]account.Account, error)
}

func NewParticipates(conf config.ParticipateConfig) (Participates, error) {
	var err error
	var participates Participates
	participatesPolicy := conf.PolicyName
	switch participatesPolicy {
	case common.SoloPolicy:
		log.Info("Get participates policy is solo.")
		participates, err = solo.NewSoloPolicy()
	case common.DposPolicy:
		log.Info("Get participates policy is dpos.")
		participates, err = dpos.NewDPOSPolicy(conf.Delegates, conf.Participates)
	default:
		log.Error("Now, we only support solo policy participates.")
		err = fmt.Errorf("not supported type")
	}
	return participates, err
}
