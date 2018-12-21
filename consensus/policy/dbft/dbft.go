package dbft

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/config"
	"github.com/DSiSc/galaxy/consensus/messages"
	"github.com/DSiSc/galaxy/consensus/utils"
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

func NewDBFTPolicy(account account.Account, timeout config.ConsensusTimeout) (*DBFTPolicy, error) {
	policy := &DBFTPolicy{
		name:    common.DbftPolicy,
		account: account,
		timeout: time.Duration(timeout.TimeoutToChangeView),
		result:  make(chan *messages.ConsensusResult),
	}
	policy.core = NewDBFTCore(account, policy.result)
	return policy, nil
}

func (instance *DBFTPolicy) Initialization(master account.Account, peers []account.Account, events types.EventCenter, onLine bool) {
	instance.core.master = master
	instance.core.commit = false
	instance.core.peers = peers
	instance.core.eventCenter = events
	instance.core.tolerance = uint8((len(peers) - 1) / 3)
	instance.core.signature = &signData{
		signatures: make([][]byte, 0),
		signMap:    make(map[account.Account][]byte),
	}
	if !onLine {
		// Add timer
		timer := time.NewTimer(30 * time.Second)
		instance.core.masterTimeout = timer
		go instance.core.waitMasterTimeOut(timer)
	}
	return
}

func (instance *DBFTPolicy) waitMasterTimeOut(timer *time.Timer) {
	for {
		select {
		case <-timer.C:
			log.Info("wait master timeout, so change view begin.")
			viewChangeReqMsg := messages.Message{
				MessageType: messages.ViewChangeMessageReqType,
				PayLoad: &messages.ViewChangeReqMessage{
					ViewChange: &messages.ViewChangeReq{
						Nodes:     []account.Account{instance.core.local},
						Timestamp: time.Now().Unix(),
						ViewNum:   instance.core.views.viewNum + 1,
					},
				},
			}
			msgRaw, err := messages.EncodeMessage(viewChangeReqMsg)
			if nil != err {
				log.Error("marshal proposal msg failed with %v.", err)
				return
			}
			messages.BroadcastPeers(msgRaw, viewChangeReqMsg.MessageType, types.Hash{}, instance.core.peers)
			return
		}
	}
}

func (instance *DBFTPolicy) PolicyName() string {
	return instance.name
}

func (instance *DBFTPolicy) Start() {
	log.Info("start dbft policy service.")
	instance.core.Start(instance.account)
}

func (instance *DBFTPolicy) commit(block *types.Block, result bool) {
	commit := &messages.Commit{
		Account:    instance.account,
		Timestamp:  time.Now().Unix(),
		Digest:     block.Header.MixDigest,
		Signatures: block.Header.SigData,
		BlockHash:  block.HeaderHash,
		Result:     result,
	}
	instance.core.SendCommit(commit, block)
}

func (instance *DBFTPolicy) ToConsensus(p *common.Proposal) error {
	var err error
	var result = false
	request := &messages.Request{
		Timestamp: p.Timestamp,
		Payload:   p.Block,
	}
	timer := time.NewTimer(time.Second * instance.timeout)
	go utils.SendEvent(instance.core, request)
	select {
	case consensusResult := <-instance.result:
		if nil != consensusResult.Result {
			log.Error("consensus for %x failed with error %v.", p.Block.Header.MixDigest, consensusResult.Result)
			err = consensusResult.Result
		} else {
			p.Block.Header.SigData = consensusResult.Signatures
			p.Block.HeaderHash = common.HeaderHash(p.Block)
			result = true
			log.Info("consensus for %x successfully with signature %x.", p.Block.Header.MixDigest, consensusResult.Signatures)
		}
		go instance.commit(p.Block, result)
		timer.Stop()
	case <-timer.C:
		log.Error("consensus for %x timeout in %d seconds.", p.Block.Header.MixDigest, instance.timeout)
		err = fmt.Errorf("timeout for consensus")
		go instance.commit(p.Block, result)
		timer.Stop()
	}
	return err
}

func (instance *DBFTPolicy) Halt() {
	return
}

func (instance *DBFTPolicy) GetConsensusResult() common.ConsensusResult {
	return common.ConsensusResult{
		View:        instance.core.views.viewNum,
		Participate: instance.core.peers,
		Master:      instance.core.master,
	}
}

func (instance *DBFTPolicy) Online() {
	return
}
