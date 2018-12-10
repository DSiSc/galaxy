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

func (self *ViewChange) GetCurrentViewNum() uint64 {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.currentView
}

func (self *ViewChange) SetCurrentViewNum(newViewNum uint64) {
	self.lock.RLock()
	self.currentView = newViewNum
	self.lock.RUnlock()
}

func (self *ViewChange) GetWalterLevel() int64 {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.walterLevel
}

func (self *ViewChange) AddViewRequest(viewNum uint64, toChange uint8) (*ViewRequests, error) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if utils.Abs(self.currentView, viewNum) > self.walterLevel {
		log.Error("diff of current view %d and request %d beyond walter level %d.",
			self.currentView, viewNum, self.walterLevel)
		return nil, fmt.Errorf("diff of current view %d and request view %d beyond walter level %d",
			self.currentView, viewNum, self.walterLevel)
	}
	if _, ok := self.viewRequests[viewNum]; !ok {
		log.Info("add view change number %d.", viewNum)
		self.viewRequests[viewNum] = NewRequests(toChange)
		return self.viewRequests[viewNum], nil
	}
	log.Warn("view sets for %d has exist and state is %v.", viewNum, self.viewRequests[viewNum].state)
	return self.viewRequests[viewNum], nil
}

func (self *ViewChange) RemoveRequest(viewNum uint64) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if _, ok := self.viewRequests[viewNum]; ok {
		log.Info("remove view change %d.", viewNum)
		delete(self.viewRequests, viewNum)
	}
	return
}

func (self *ViewChange) GetRequestByViewNum(viewNum uint64) *ViewRequests {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.viewRequests[viewNum]
}

func (self *ViewRequests) ReceiveViewRequestByAccount(account account.Account) ViewRequestState {
	self.lock.Lock()
	defer self.lock.Unlock()
	if _, ok := self.index[account]; !ok {
		log.Info("add %x view request.", account.Address)
		self.index[account] = true
		self.nodes = append(self.nodes, account)
		if uint8(len(self.nodes)) >= self.toChange {
			log.Info("request has reach to change view situation which need less than %d, now received is %d.", len(self.nodes), self.toChange)
			self.state = ViewEnd
		}
	} else {
		log.Warn("has receive %x view request.", account.Address)
	}
	return self.state
}

func (self *ViewRequests) GetViewRequestState() ViewRequestState {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.state
}

func (self *ViewRequests) GetReceivedAccounts() []account.Account {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.nodes
}
