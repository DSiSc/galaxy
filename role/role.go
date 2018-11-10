package role

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/galaxy/role/config"
	"github.com/DSiSc/galaxy/role/policy/dpos"
	"github.com/DSiSc/galaxy/role/policy/solo"
	"github.com/DSiSc/validator/tools/account"
)

type Role interface {
	PolicyName() string
	RoleAssignments() (map[account.Account]common.Roler, error)
	GetRoles(address account.Account) (common.Roler, error)
}

func NewRole(participates participates.Participates, conf config.RoleConfig) (Role, error) {
	var err error
	var role Role
	rolePolicy := conf.PolicyName
	switch rolePolicy {
	case common.SOLO_POLICY:
		log.Info("Get role policy is solo.")
		role, err = solo.NewSoloPolicy(participates)
	case common.DPOS_POLICY:
		log.Info("Get role policy is dpos.")
		role, err = dpos.NewDPOSPolicy(participates)
	default:
		log.Error("Now, we only support solo role policy.")
		err = fmt.Errorf("unkonwn policy type")
	}
	return role, err
}
