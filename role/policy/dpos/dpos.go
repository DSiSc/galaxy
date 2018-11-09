package dpos

import (
	"fmt"
	"github.com/DSiSc/blockchain"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/validator/tools/account"
)

type rolesAssignment struct {
	delegates []account.Account
	roles     map[account.Account]common.Roler
}

type DPOSPolicy struct {
	name        string
	local       account.Account
	participate participates.Participates
	roles       rolesAssignment
}

func NewDPOSPolicy(p participates.Participates, localNode account.Account) (*DPOSPolicy, error) {
	policy := &DPOSPolicy{
		name:        common.DPOS_POLICY,
		local:       localNode,
		participate: p,
	}
	return policy, nil
}

func (self *DPOSPolicy) RoleAssignments() (map[account.Account]common.Roler, error) {
	// TODO: simply we decide master by block height, while it will support appoint by external
	delegates, err := self.participate.GetParticipates()
	if err != nil {
		log.Error("Error to get participates.")
		return nil, fmt.Errorf("get participates with error: %s", err)
	}

	block, ok := blockchain.NewLatestStateBlockChain()
	if nil != ok {
		log.Error("Get NewLatestStateBlockChain failed.")
		return nil, fmt.Errorf("get NewLatestStateBlockChain failed")
	}

	numberOfDelegates := len(delegates)
	assignment := make(map[account.Account]common.Roler, numberOfDelegates)

	currentBlockHeight := block.GetCurrentBlock().Header.Height
	masterIndex := (currentBlockHeight + 1) % uint64(numberOfDelegates)

	for index, delegate := range delegates {
		if index == int(masterIndex) {
			assignment[delegate] = common.Master
		} else {
			assignment[delegate] = common.Slave
		}
	}
	self.roles = rolesAssignment{
		delegates: delegates,
		roles:     assignment,
	}

	return assignment, nil
}

func (self *DPOSPolicy) GetRoles(address account.Account) common.Roler {
	if role, ok := self.roles.roles[address]; !ok {
		log.Error("account %x is not a delegate, please confirm.", address)
		// TODO: verify normal node or unknown
		return common.UnKnown
	} else {
		log.Info("account %x role is %v.", address, role)
		return role
	}
}

func (self *DPOSPolicy) PolicyName() string {
	return self.name
}

func (self *DPOSPolicy) AppointRole(master int) {
	if master > len(self.roles.delegates) {
		log.Error("master index %d exceed delegates span %d.", master, len(self.roles.delegates))
		return
	}

	masterAccount := self.roles.delegates[master]
	if _, ok := self.roles.roles[masterAccount]; !ok {
		log.Error("account %x has not assign role, please confirm.", masterAccount.Address)
		return
	}

	for delegate, role := range self.roles.roles {
		if delegate == masterAccount {
			if common.Master == role {
				log.Warn("Account %x has been master already.", masterAccount.Address)
				return
			} else {
				self.roles.roles[masterAccount] = common.Master
			}
		} else {
			self.roles.roles[delegate] = common.Slave
		}
	}
}
