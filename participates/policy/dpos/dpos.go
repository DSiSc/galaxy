package dpos

import (
	"fmt"
	"github.com/DSiSc/contractsManage/contracts"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/galaxy/participates/common"
	"github.com/DSiSc/validator/tools/account"
)

type DPOSPolicy struct {
	name     string
	contract contracts.Voting
	// number of delegates, update when called GetParticipates
	members      uint64
	participates []account.Account
}

func NewDPOSPolicy() *DPOSPolicy {
	voting := contracts.NewVotingContract()
	return &DPOSPolicy{
		name:     common.DposPolicy,
		contract: voting,
	}
}

func (instance *DPOSPolicy) PolicyName() string {
	return instance.name
}

// Get the top ranking of count from voting result.
func (instance *DPOSPolicy) getDelegatesByCount(count uint64) ([]account.Account, error) {
	// TODO: Get accounts by voting result
	var accounts = make([]account.Account, 0)
	nodeList, err := instance.contract.GetNodeList(count)
	if nil != err {
		log.Error("get node list failed with %v", err)
		return accounts, fmt.Errorf("get node list failed with %v", err)
	}
	for _, node := range nodeList {
		account := account.Account{
			Address: node.Address,
			Extension: account.AccountExtension{
				Id:  node.Id,
				Url: node.Url,
			},
		}
		accounts = append(accounts, account)
	}
	instance.participates = accounts
	return instance.participates, nil
}

func (instance *DPOSPolicy) getDelegates() ([]account.Account, error) {
	instance.members = instance.contract.NodeNumber()
	delegates, err := instance.getDelegatesByCount(instance.members)
	if nil != err {
		log.Error("get delegates failed with err %s.", err)
		return nil, err
	}
	return delegates, nil
}

func (instance *DPOSPolicy) GetParticipates() ([]account.Account, error) {
	participates, err := instance.getDelegates()
	if nil != err {
		log.Error("Get delegates failed with error %v.", err)
	} else {
		instance.participates = participates
	}
	return participates, err
}
