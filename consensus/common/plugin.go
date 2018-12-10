package common

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/validator/tools/account"
	"sync"
)

type contentState uint8

const (
	NullState = contentState(iota)
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
