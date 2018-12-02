package dbft

import (
	"encoding/json"
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/policy/bft/messages"
	"github.com/DSiSc/galaxy/consensus/policy/bft/tools"
	commonr "github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/validator/tools/account"
	"time"
)

type DBFTPolicy struct {
	name string
	// local account
	account account.Account
	core    *dbftCore
	timeout time.Duration
	result  chan *messages.ConsensusResult
}

func NewDBFTPolicy(account account.Account, timeout int64) (*DBFTPolicy, error) {
	policy := &DBFTPolicy{
		name:    common.DBFT_POLICY,
		account: account,
		timeout: time.Duration(timeout),
		result:  make(chan *messages.ConsensusResult),
	}
	policy.core = NewDBFTCore(account, policy.result)
	return policy, nil
}

func (self *DBFTPolicy) Initialization(role map[account.Account]commonr.Roler, peers []account.Account, events types.EventCenter) error {
	if len(role) != len(peers) {
		log.Error("dbft core has not been initial, please confirm.")
		return fmt.Errorf("role and peers not in consistent")
	}
	var masterExist bool = false
	for delegate, role := range role {
		if commonr.Master == role {
			self.core.master = delegate.Extension.Id
			masterExist = true
			break
		}
	}
	if !masterExist {
		log.Error("no master exist in delegates")
		return fmt.Errorf("no master")
	}
	self.core.commit = false
	self.core.peers = peers
	self.core.eventCenter = events
	self.core.tolerance = uint8((len(peers) - 1) / 3)
	self.core.signature = &signData{
		signatures: make([][]byte, 0),
		signMap:    make(map[account.Account][]byte),
	}
	// Add timer
	timer := time.NewTimer(10 * time.Second)
	go self.waitMasterTimeOut(timer)
	return nil
}

func (self *DBFTPolicy) waitMasterTimeOut(timer *time.Timer) {
	for {
		select {
		case <-timer.C:
			log.Info("wait master timeout, so change view begin.")
			viewChangeReqMsg := &messages.Message{
				MessageType: messages.ViewChangeMessageReqType,
				Payload: &messages.ViewChangeReqMessage{
					ViewChange: &messages.ViewChangeReq{
						Nodes:     []account.Account{self.core.local},
						Timestamp: time.Now().Unix(),
						ViewNum:   self.core.views.viewNum + 1,
					},
				},
			}
			msgRaw, err := json.Marshal(viewChangeReqMsg)
			if nil != err {
				log.Error("marshal proposal msg failed with %v.", err)
				return
			}
			messages.BroadcastPeers(msgRaw, messages.ViewChangeMessageReqType, types.Hash{}, self.core.peers)
			return
		}
	}
}

func (self *DBFTPolicy) PolicyName() string {
	return self.name
}

func (self *DBFTPolicy) Start() {
	log.Info("start bft policy service.")
	self.core.Start(self.account)
}

func (self *DBFTPolicy) commit(block *types.Block, result bool) {
	commit := &messages.Commit{
		Account:    self.account,
		Timestamp:  time.Now().Unix(),
		Digest:     block.Header.MixDigest,
		Signatures: block.Header.SigData,
		BlockHash:  block.HeaderHash,
		Result:     result,
	}
	self.core.SendCommit(commit, block)
}

func (self *DBFTPolicy) ToConsensus(p *common.Proposal) error {
	var err error
	var result = false
	request := &messages.Request{
		Timestamp: p.Timestamp,
		Payload:   p.Block,
	}
	timer := time.NewTimer(time.Second * self.timeout)
	go tools.SendEvent(self.core, request)
	select {
	case consensusResult := <-self.result:
		if nil != consensusResult.Result {
			log.Error("consensus for %x failed with error %v.", p.Block.Header.MixDigest, consensusResult.Result)
			err = consensusResult.Result
		} else {
			p.Block.Header.SigData = consensusResult.Signatures
			p.Block.HeaderHash = common.HeaderHash(p.Block)
			result = true
			log.Info("consensus for %x successfully with signature %x.", p.Block.Header.MixDigest, consensusResult.Signatures)
		}
		go self.commit(p.Block, result)
		timer.Stop()
	case <-timer.C:
		log.Error("consensus for %x timeout in %d seconds.", p.Block.Header.MixDigest, self.timeout)
		err = fmt.Errorf("timeout for consensus")
		go self.commit(p.Block, result)
		timer.Stop()
	}
	return err
}

func (self *DBFTPolicy) Halt() {
	return
}

func (self *DBFTPolicy) GetConsensusResult() common.ConsensusResult {
	var assignment = make(map[account.Account]commonr.Roler)
	for _, node := range self.core.peers {
		if node.Extension.Id == self.core.master {
			assignment[node] = commonr.Master
			continue
		}
		assignment[node] = commonr.Slave
	}
	return common.ConsensusResult{
		View:        self.core.views.viewNum,
		Participate: self.core.peers,
		Roles:       assignment,
	}
}
