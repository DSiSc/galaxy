package tools

import (
	"bytes"
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/validator/tools/account"
	"sync"
)

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

type ConsensusMap struct {
	mutex        sync.RWMutex
	consensusMap map[types.Hash]*ConsensusContent
}

func NewConsensusMap() *ConsensusMap {
	return &ConsensusMap{
		consensusMap: make(map[types.Hash]*ConsensusContent),
	}
}

func (self *ConsensusMap) GetConsensusMap() map[types.Hash]*ConsensusContent {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.consensusMap
}

func (self *ConsensusMap) GetConsensusContentByHash(digest types.Hash) *ConsensusContent {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.consensusMap[digest]
}

func (self *ConsensusMap) GetConsensusContentSigMapByHash(digest types.Hash) (map[account.Account][]byte, error) {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	if _, ok := self.consensusMap[digest]; !ok {
		log.Info("record %x not exist, please confirm.", digest)
		self.consensusMap[digest] = NewConsensusContent(digest)
		return nil, fmt.Errorf("item %x not exist", digest)
	}

	return self.consensusMap[digest].GetSignMapByHash(digest)
}

func (self *ConsensusMap) Add(digest types.Hash) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	if _, ok := self.consensusMap[digest]; !ok {
		log.Info("add item %x to map to prepare consensus process.", digest)
		self.consensusMap[digest] = NewConsensusContent(digest)
		return
	}
	log.Warn("item %x has exist, please confirm.", digest)
	return
}

func (self *ConsensusMap) Remove(digest types.Hash) {
	self.mutex.Lock()
	delete(self.consensusMap, digest)
	self.mutex.Unlock()
}

type consensusState uint8

const (
	NIL = consensusState(iota)
	Initial
	InConsensus
	ToConsensus
)

type ConsensusContent struct {
	digest     types.Hash
	state      consensusState
	lock       sync.RWMutex
	signMap    map[account.Account][]byte
	signatures [][]byte
	content    interface{}
}

func NewConsensusContent(digest types.Hash) *ConsensusContent {
	return &ConsensusContent{
		digest:  digest,
		state:   Initial,
		signMap: make(map[account.Account][]byte),
	}
}

func (self *ConsensusContent) SetState(state consensusState) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if 1 != (state - self.state) {
		log.Error("can not move state from %v to %v.", self.state, state)
		return
	}
	log.Debug("set state from %v to %v.")
	self.state = state
	return
}

func (self *ConsensusContent) AddSignature(account account.Account, sign []byte) bool {
	self.lock.Lock()
	defer self.lock.Unlock()
	if _, ok := self.signMap[account]; !ok {
		log.Info("add account %d signature.", account.Extension.Id)
		self.signMap[account] = sign
		self.signatures = append(self.signatures, sign)
		return true
	}
	log.Warn("account %d signature has exist.", account.Extension.Id)
	return false
}

func (self *ConsensusContent) Signatures() int {
	self.lock.Lock()
	defer self.lock.Unlock()
	return len(self.signatures)
}

func (self *ConsensusContent) GetSignByAccount(account account.Account) (bool, []byte) {
	self.lock.Lock()
	defer self.lock.Unlock()
	sign, ok := self.signMap[account]
	return ok, sign
}

func (self *ConsensusContent) GetSignMapByHash(digest types.Hash) (map[account.Account][]byte, error) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if !bytes.Equal(self.digest[:], digest[:]) {
		log.Error("wrong digest which expect is %v while setting is %v.", self.digest, digest)
		return nil, fmt.Errorf("record not exist with digest %v", digest)
	}
	return self.signMap, nil
}

func (self *ConsensusContent) SetContentByHash(digest types.Hash, payload interface{}) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if !bytes.Equal(self.digest[:], digest[:]) {
		log.Error("wrong digest which expect is %v while setting is %v.", self.digest, digest)
		return
	}
	self.content = payload
	return
}

func (self *ConsensusContent) GetContentByHash(digest types.Hash) (interface{}, error) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if !bytes.Equal(self.digest[:], digest[:]) {
		log.Error("wrong digest which expect is %v while setting is %v.", self.digest, digest)
		return nil, fmt.Errorf("record not exist with digest %v", digest)
	}
	return self.content, nil
}
