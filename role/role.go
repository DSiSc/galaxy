package role

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/galaxy/role/config"
	"github.com/DSiSc/galaxy/role/policy/dpos"
	"github.com/DSiSc/galaxy/role/policy/solo"
	"github.com/DSiSc/validator/tools/account"
)

type Role interface {
	PolicyName() string
	RoleAssignments(accounts []account.Account) (map[account.Account]common.Roler, error)
	ChangeRoleAssignment(map[account.Account]common.Roler, uint64)
	GetRoles(address account.Account) (common.Roler, error)
}

func NewRole(conf config.RoleConfig) (Role, error) {
	var err error
	var role Role
	rolePolicy := conf.PolicyName
	switch rolePolicy {
	case common.SOLO_POLICY:
		log.Info("Get role policy is solo.")
		role, err = solo.NewSoloPolicy()
	case common.DPOS_POLICY:
		log.Info("Get role policy is dpos.")
		role, err = dpos.NewDPOSPolicy()
	default:
		log.Error("Now, we only support solo role policy.")
		err = fmt.Errorf("unkonwn policy type")
	}
	return role, err
}
