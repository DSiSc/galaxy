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
	_ = contentState(iota)
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

func (instance *Content) SetState(state contentState) error {
	instance.lock.Lock()
	if 1 != (state - instance.state) {
		instance.lock.Unlock()
		return fmt.Errorf("can not move state from %v to %v", instance.state, state)
	}
	log.Debug("set state from %v to %v.", instance.state, state)
	instance.state = state
	instance.lock.Unlock()
	return nil
}

func (instance *Content) AddSignature(account account.Account, sign []byte) bool {
	instance.lock.Lock()
	if _, ok := instance.signMap[account]; !ok {
		log.Info("add account %d signature.", account.Extension.Id)
		instance.signMap[account] = sign
		instance.signatures = append(instance.signatures, sign)
		instance.lock.Unlock()
		return true
	}
	instance.lock.Unlock()
	log.Warn("account %d signature has exist.", account.Extension.Id)
	return false
}

func (instance *Content) Signatures() [][]byte {
	instance.lock.RLock()
	defer instance.lock.RUnlock()
	return instance.signatures
}

func (instance *Content) State() contentState {
	instance.lock.RLock()
	defer instance.lock.RUnlock()
	return instance.state
}

func (instance *Content) GetSignByAccount(account account.Account) ([]byte, bool) {
	instance.lock.RLock()
	sign, ok := instance.signMap[account]
	instance.lock.RUnlock()
	return sign, ok
}

func (instance *Content) GetSignMap() map[account.Account][]byte {
	instance.lock.RLock()
	defer instance.lock.RUnlock()
	return instance.signMap
}

func (instance *Content) GetContentPayload() interface{} {
	instance.lock.RLock()
	defer instance.lock.RUnlock()
	return instance.payload
}

type ConsensusPlugin struct {
	mutex             sync.RWMutex
	latestBlockHeight uint64
	content           map[types.Hash]*Content
}

func NewConsensusPlugin() *ConsensusPlugin {
	return &ConsensusPlugin{
		latestBlockHeight: uint64(0),
		content:           make(map[types.Hash]*Content),
	}
}

func (instance *ConsensusPlugin) Add(digest types.Hash, payload interface{}) *Content {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	if _, ok := instance.content[digest]; !ok {
		log.Info("add content %x to map to prepare consensus process.", digest)
		instance.content[digest] = newContent(digest, payload)
		return instance.content[digest]
	}
	log.Warn("content %x has exist, please confirm.", digest)
	return instance.content[digest]
}

func (instance *ConsensusPlugin) Remove(digest types.Hash) {
	instance.mutex.Lock()
	delete(instance.content, digest)
	instance.mutex.Unlock()
}

func (instance *ConsensusPlugin) GetContentByHash(digest types.Hash) (*Content, error) {
	instance.mutex.RLock()
	defer instance.mutex.RUnlock()
	if _, ok := instance.content[digest]; !ok {
		log.Error("content %x not exist, please confirm.", digest)
		return nil, fmt.Errorf("content %x not exist, please confirm", digest)
	}
	return instance.content[digest], nil
}

func (instance *ConsensusPlugin) GetLatestBlockHeight() uint64 {
	instance.mutex.RLock()
	defer instance.mutex.RUnlock()
	return instance.latestBlockHeight
}

func (instance *ConsensusPlugin) SetLatestBlockHeight(height uint64) {
	instance.mutex.Lock()
	instance.latestBlockHeight = height
	instance.mutex.Unlock()
}

type OnlineWizard struct {
	mutex           sync.RWMutex
	blockHeight     uint64
	blockHeightList []uint64
	response        map[uint64]*ResponseNodes
}

type ResponseNodes struct {
	mutex       sync.RWMutex
	nodes       []account.Account
	node        map[types.Address]account.Account
	state       OnlineState
	master      account.Account
	viewNum     uint64
	walterLevel int
}

func NewResponseNodes(walterLevel int, master account.Account) *ResponseNodes {
	return &ResponseNodes{
		node:        make(map[types.Address]account.Account),
		state:       GoOnline,
		master:      master,
		walterLevel: walterLevel,
		nodes:       make([]account.Account, 0),
	}
}

func (instance *ResponseNodes) GetResponseState() OnlineState {
	instance.mutex.RLock()
	defer instance.mutex.RUnlock()
	return instance.state
}

func (instance *ResponseNodes) AddResponseNodes(node account.Account) ([]account.Account, OnlineState) {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	if _, ok := instance.node[node.Address]; !ok {
		instance.node[node.Address] = node
		instance.nodes = append(instance.nodes, node)
	}
	if len(instance.nodes) >= instance.walterLevel {
		log.Info("now node has reached consensus, so state from GoOnline to Online")
		instance.state = Online
	}
	return instance.nodes, instance.state
}

func NewOnlineWizard() *OnlineWizard {
	return &OnlineWizard{
		blockHeight:     DefaultBlockHeight,
		blockHeightList: make([]uint64, 0),
		response:        make(map[uint64]*ResponseNodes),
	}
}

func (instance *OnlineWizard) AddOnlineResponse(blockHeight uint64, nodes []account.Account, walterLevel int, master account.Account, viewNum uint64) ([]account.Account, OnlineState) {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	var nodeList []account.Account
	var state OnlineState
	if _, ok := instance.response[blockHeight]; !ok {
		instance.response[blockHeight] = NewResponseNodes(walterLevel, master)
		instance.blockHeightList = append(instance.blockHeightList, blockHeight)
	}
	if blockHeight > instance.blockHeight {
		log.Info("update block height from %d to %d.", instance.blockHeight, blockHeight)
		instance.blockHeight = blockHeight
		instance.response[blockHeight].master = master
	}
	if master != instance.response[blockHeight].master {
		log.Error("master not in agreement which exist is %d, while receive is %d.",
			instance.response[blockHeight].master.Extension.Id, master.Extension.Id)
		if instance.response[blockHeight].viewNum < viewNum {
			instance.response[blockHeight].master = master
		} else {
			panic(fmt.Sprintf("master not in agreement which exist is %d, while receive is %d.",
				instance.response[blockHeight].master.Extension.Id, master.Extension.Id))
		}
	}
	if instance.response[blockHeight].viewNum < viewNum {
		instance.response[blockHeight].viewNum = viewNum
	}
	for _, node := range nodes {
		nodeList, state = instance.response[blockHeight].AddResponseNodes(node)
	}
	return nodeList, state
}

func (instance *OnlineWizard) GetCurrentState() OnlineState {
	instance.mutex.RLock()
	defer instance.mutex.RUnlock()
	return instance.response[instance.blockHeight].GetResponseState()
}

func (instance *OnlineWizard) GetCurrentStateByHeight(blockHeight uint64) OnlineState {
	instance.mutex.RLock()
	defer instance.mutex.RUnlock()
	if _, ok := instance.response[blockHeight]; !ok {
		log.Debug("block height %d info not exists.", blockHeight)
		return GoOnline
	}
	return instance.response[blockHeight].GetResponseState()
}

func (instance *OnlineWizard) GetCurrentHeight() uint64 {
	instance.mutex.RLock()
	defer instance.mutex.RUnlock()
	return instance.blockHeight
}

func (instance *OnlineWizard) GetMasterByBlockHeight(blockHeight uint64) account.Account {
	instance.mutex.RLock()
	defer instance.mutex.RUnlock()
	return instance.response[blockHeight].master
}

func (instance *OnlineWizard) GetResponseNodesStateByBlockHeight(blockHeight uint64) OnlineState {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	if _, ok := instance.response[blockHeight]; !ok {
		log.Warn("never received %d response before.", blockHeight)
		return GoOnline
	}
	return instance.response[blockHeight].state
}

func (instance *OnlineWizard) DeleteOnlineResponse(blockHeight uint64) {
	instance.mutex.Lock()
	delete(instance.response, blockHeight)
	instance.mutex.Unlock()
}
