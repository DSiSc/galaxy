package role

import (
	"github.com/DSiSc/galaxy/participates"
	rolec "github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/galaxy/role/policy"
	"github.com/DSiSc/producer/config"
	"github.com/DSiSc/txpool/common"
	"github.com/DSiSc/txpool/common/log"
)

type Role interface {
	PolicyName() string
	RoleAssignments() (map[common.Address]rolec.Roler, error)
	GetRoles(address common.Address) rolec.Roler
}

const (
	ROLE_SOLO = "solo"
	// Structure must matching with defination of config/config.json
	Symbol = "role"
	Policy = "role.policy"
)

func NewRolePolicy(p participates.Participates, address common.Address) (Role, error) {
	var err error
	var role Role
	conf := config.New(config.ConfigAbsPath())
	rolePolicy := conf.GetConfigItem(Policy).(string)
	switch rolePolicy {
	case ROLE_SOLO:
		log.Info("Get role policy is solo.")
		role, err = policy.NewSoloPolicy(p, address)
	default:
		log.Error("Now, we only support solo role policy.")
	}
	return role, err
}
