package common

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/galaxy/consensus/utils"
	"github.com/DSiSc/validator/tools/account"
	"sync"
)

type ViewChange struct {
	lock         sync.RWMutex
	currentView  uint64
	walterLevel  int64
	viewRequests map[uint64]*ViewRequests
}

type ViewRequests struct {
	lock     sync.RWMutex
	toChange uint8
	state    ViewRequestState
	index    map[account.Account]bool
	nodes    []account.Account
}

func NewRequests(toChange uint8) *ViewRequests {
	return &ViewRequests{
		state:    Viewing,
		toChange: toChange,
		index:    make(map[account.Account]bool),
		nodes:    make([]account.Account, 0),
	}
}

func NewViewChange() *ViewChange {
	return &ViewChange{
		currentView:  DefaultViewNum,
		walterLevel:  DefaultWalterLevel,
		viewRequests: make(map[uint64]*ViewRequests),
	}
}

func (instance *ViewChange) GetCurrentViewNum() uint64 {
	instance.lock.RLock()
	defer instance.lock.RUnlock()
	return instance.currentView
}

func (instance *ViewChange) SetCurrentViewNum(newViewNum uint64) {
	instance.lock.RLock()
	instance.currentView = newViewNum
	instance.lock.RUnlock()
}

func (instance *ViewChange) GetWalterLevel() int64 {
	instance.lock.RLock()
	defer instance.lock.RUnlock()
	return instance.walterLevel
}

func (instance *ViewChange) AddViewRequest(viewNum uint64, toChange uint8) (*ViewRequests, error) {
	instance.lock.Lock()
	defer instance.lock.Unlock()
	if utils.Abs(instance.currentView, viewNum) > instance.walterLevel {
		log.Error("diff of current view %d and request %d beyond walter level %d.",
			instance.currentView, viewNum, instance.walterLevel)
		return nil, fmt.Errorf("diff of current view %d and request view %d beyond walter level %d",
			instance.currentView, viewNum, instance.walterLevel)
	}
	if _, ok := instance.viewRequests[viewNum]; !ok {
		log.Info("add view change number %d.", viewNum)
		instance.viewRequests[viewNum] = NewRequests(toChange)
		return instance.viewRequests[viewNum], nil
	}
	log.Warn("view sets for %d has exist and state is %v.", viewNum, instance.viewRequests[viewNum].state)
	return instance.viewRequests[viewNum], nil
}

func (instance *ViewChange) RemoveRequest(viewNum uint64) {
	instance.lock.Lock()
	defer instance.lock.Unlock()
	if _, ok := instance.viewRequests[viewNum]; ok {
		log.Info("remove view change %d.", viewNum)
		delete(instance.viewRequests, viewNum)
	}
	return
}

func (instance *ViewChange) GetRequestByViewNum(viewNum uint64) *ViewRequests {
	instance.lock.RLock()
	defer instance.lock.RUnlock()
	return instance.viewRequests[viewNum]
}

func (instance *ViewRequests) ReceiveViewRequestByAccount(account account.Account) ViewRequestState {
	instance.lock.Lock()
	defer instance.lock.Unlock()
	if _, ok := instance.index[account]; !ok {
		log.Info("add %x view request.", account.Address)
		instance.index[account] = true
		instance.nodes = append(instance.nodes, account)
		if uint8(len(instance.nodes)) >= instance.toChange {
			log.Info("request has reach to change view situation which need less than %d, now received is %d.", len(instance.nodes), instance.toChange)
			instance.state = ViewEnd
		}
	} else {
		log.Warn("has receive %x view request.", account.Address)
	}
	return instance.state
}

func (instance *ViewRequests) GetViewRequestState() ViewRequestState {
	instance.lock.RLock()
	defer instance.lock.RUnlock()
	return instance.state
}

func (instance *ViewRequests) GetReceivedAccounts() []account.Account {
	instance.lock.RLock()
	defer instance.lock.RUnlock()
	return instance.nodes
}
