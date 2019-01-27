package dpos

import (
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/galaxy/participates/common"
	"github.com/DSiSc/validator/tools/account"
)

type DPOSPolicy struct {
	name string
	// number of delegates
	members      uint64
	participates []account.Account
}

func NewDPOSPolicy(number uint64, participates []account.Account) (*DPOSPolicy, error) {
	return &DPOSPolicy{
		name:         common.DposPolicy,
		members:      number,
		participates: participates,
	}, nil
}

func (instance *DPOSPolicy) PolicyName() string {
	return instance.name
}

// Get the top ranking of count from voting result.
func (instance *DPOSPolicy) getDelegatesByCount(count uint64) ([]account.Account, error) {
	// TODO: Get accounts by voting result
	return instance.participates, nil
}

func (instance *DPOSPolicy) getDelegates() ([]account.Account, error) {
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
