package tools

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/validator/tools/account"
	"sync"
)

func Abs(x uint64, y uint64) int64 {
	n := int64(x - y)
	z := n >> 63
	return (n ^ z) - z
}

func AccountFilter(blacklist []account.Account, accounts []account.Account) []account.Account {
	var peer []account.Account
	for _, black := range blacklist {
		peer = filterAccount(black, accounts)
	}
	return peer
}

func filterAccount(black account.Account, accounts []account.Account) []account.Account {
	all := make([]account.Account, 0)
	for _, account := range accounts {
		if black != account {
			all = append(all, account)
		}
	}
	return all
}

func GetAccountById(peers []account.Account, except uint64) account.Account {
	var temp account.Account
	for _, peer := range peers {
		if peer.Extension.Id == except {
			temp = peer
		}
	}
	return temp
}

func GetNodeAccountWithMinId(nodes []account.Account) account.Account {
	var minNode = nodes[0]
	for _, node := range nodes {
		if node.Extension.Id < minNode.Extension.Id {
			minNode = node
		}
	}
	return minNode
}

type contentState uint8

const (
	NIL = contentState(iota)
	Initial
	InConsensus
	ToConsensus
)

type Content struct {
	digest     types.Hash
	state      contentState
	lock       sync.RWMutex
	signMap    map[account.Account][]byte
	signatures [][]byte
	payload    interface{}
}

func newContent(digest types.Hash, payload interface{}) *Content {
	return &Content{
		digest:  digest,
		state:   Initial,
		payload: payload,
		signMap: make(map[account.Account][]byte),
	}
}

func (self *Content) SetState(state contentState) error {
	self.lock.Lock()
	if 1 != (state - self.state) {
		self.lock.Unlock()
		return fmt.Errorf("can not move state from %v to %v", self.state, state)
	}
	log.Debug("set state from %v to %v.", self.state, state)
	self.state = state
	self.lock.Unlock()
	return nil
}

func (self *Content) AddSignature(account account.Account, sign []byte) bool {
	self.lock.Lock()
	if _, ok := self.signMap[account]; !ok {
		log.Info("add account %d signature.", account.Extension.Id)
		self.signMap[account] = sign
		self.signatures = append(self.signatures, sign)
		self.lock.Unlock()
		return true
	}
	self.lock.Unlock()
	log.Warn("account %d signature has exist.", account.Extension.Id)
	return false
}

func (self *Content) Signatures() [][]byte {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.signatures
}

func (self *Content) State() contentState {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.state
}

func (self *Content) GetSignByAccount(account account.Account) ([]byte, bool) {
	self.lock.RLock()
	sign, ok := self.signMap[account]
	self.lock.RUnlock()
	return sign, ok
}

func (self *Content) GetSignMap() map[account.Account][]byte {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.signMap
}

func (self *Content) GetContentPayloadByHash(digest types.Hash) interface{} {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.payload
}

type ConsensusPlugin struct {
	mutex   sync.RWMutex
	content map[types.Hash]*Content
}

func NewConsensusPlugin() *ConsensusPlugin {
	return &ConsensusPlugin{
		content: make(map[types.Hash]*Content),
	}
}

func (self *ConsensusPlugin) Add(digest types.Hash, payload interface{}) *Content {
	self.mutex.Lock()
	if _, ok := self.content[digest]; !ok {
		log.Info("add content %x to map to prepare consensus process.", digest)
		self.content[digest] = newContent(digest, payload)
	}
	self.mutex.Unlock()
	log.Warn("content %x has exist, please confirm.", digest)
	return self.content[digest]
}

func (self *ConsensusPlugin) Remove(digest types.Hash) {
	self.mutex.Lock()
	delete(self.content, digest)
	self.mutex.Unlock()
}

func (self *ConsensusPlugin) GetContentByHash(digest types.Hash) (*Content, error) {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	if _, ok := self.content[digest]; !ok {
		log.Error("content %x not exist, please confirm.", digest)
		return nil, fmt.Errorf("content %x not exist, please confirm", digest)
	}
	return self.content[digest], nil
}

type ViewChange struct {
	lock         sync.RWMutex
	currentView  uint64
	walterLevel  int64
	viewRequests map[uint64]*ViewRequests
}

type ViewRequests struct {
	lock     sync.RWMutex
	toChange uint8
	state    common.ViewRequestState
	index    map[account.Account]bool
	nodes    []account.Account
}

func NewRequests(toChange uint8) *ViewRequests {
	return &ViewRequests{
		state:    common.Viewing,
		toChange: toChange,
		index:    make(map[account.Account]bool),
		nodes:    make([]account.Account, 0),
	}
}

func NewViewChange() *ViewChange {
	return &ViewChange{
		currentView:  common.DefaultViewNum,
		walterLevel:  common.DefaultWalterLevel,
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
	if Abs(self.currentView, viewNum) > self.walterLevel {
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

func (self *ViewRequests) ReceiveViewRequestByAccount(account account.Account) common.ViewRequestState {
	self.lock.Lock()
	defer self.lock.Unlock()
	if _, ok := self.index[account]; !ok {
		log.Info("add %x view request.", account.Address)
		self.index[account] = true
		self.nodes = append(self.nodes, account)
		if uint8(len(self.nodes)) >= self.toChange {
			log.Info("request has reach to change view situation which need less than %d, now received is %d.", len(self.nodes), self.toChange)
			self.state = common.ViewEnd
		}
	} else {
		log.Warn("has receive %x view request.", account.Address)
	}
	return self.state
}

func (self *ViewRequests) GetViewRequestState() common.ViewRequestState {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.state
}

func (self *ViewRequests) GetReceivedAccounts() []account.Account {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.nodes
}
