package fbft

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/config"
	"github.com/DSiSc/galaxy/consensus/messages"
	"github.com/DSiSc/galaxy/consensus/utils"
	"github.com/DSiSc/repository"
	"github.com/DSiSc/validator/tools/account"
	"net"
	"strconv"
	"sync/atomic"
	"time"
)

type nodesInfo struct {
	local  account.Account
	master account.Account
	peers  []account.Account
}

func newNodesInfo(local account.Account, master account.Account, peers []account.Account) *nodesInfo {
	return &nodesInfo{
		local:  local,
		master: master,
		peers:  peers,
	}
}

func (self *nodesInfo) clone() *nodesInfo {
	info := &nodesInfo{
		local:  self.local,
		master: self.master,
		peers:  self.peers,
	}
	return info
}

type coreTimeout struct {
	timeToCollectResponseMsg int64
	timeToWaitCommitMsg      int64
	timeToChangeViewTime     int64
	timeToChangeViewStopChan chan struct{}
	isRunning                int32
}

const (
	blockSyncReqCacheLimit = 40
	connReadTimeOut        = 60
)

type blockSyncRequest struct {
	start  uint64
	end    uint64
	target account.Account
}

type fbftCore struct {
	nodes                      atomic.Value
	tolerance                  atomic.Value
	status                     common.ViewStatus
	coreTimer                  coreTimeout
	result                     chan messages.ConsensusResult
	signal                     chan common.MessageSignal
	online                     chan messages.OnlineResponse
	onlineWizard               *common.OnlineWizard
	eventCenter                types.EventCenter
	blockSwitch                chan<- interface{}
	consensusPlugin            *common.ConsensusPlugin
	viewChange                 *common.ViewChange
	enableEmptyBlock           bool
	enableSyncVerifySignature  bool
	enableLocalVerifySignature bool
	blockSyncChan              chan *blockSyncRequest
}

func NewFBFTCore(blockSwitch chan<- interface{}, timer config.ConsensusTimeout, emptyBlock bool, signatureVerify config.SignatureVerifySwitch) *fbftCore {
	instance := &fbftCore{
		enableEmptyBlock: emptyBlock,
		blockSwitch:      blockSwitch,
		status:           common.ViewNormal,
		viewChange:       common.NewViewChange(),
		onlineWizard:     common.NewOnlineWizard(),
		consensusPlugin:  common.NewConsensusPlugin(),
		signal:           make(chan common.MessageSignal),
		online:           make(chan messages.OnlineResponse),
		result:           make(chan messages.ConsensusResult),
		enableSyncVerifySignature:  signatureVerify.SyncVerifySignature,
		enableLocalVerifySignature: signatureVerify.LocalVerifySignature,
		coreTimer: coreTimeout{
			timeToCollectResponseMsg: timer.TimeoutToCollectResponseMsg,
			timeToWaitCommitMsg:      timer.TimeoutToWaitCommitMsg,
			timeToChangeViewTime:     timer.TimeoutToChangeView,
			timeToChangeViewStopChan: make(chan struct{}),
		},
		blockSyncChan: make(chan *blockSyncRequest, blockSyncReqCacheLimit),
	}
	instance.tolerance.Store(uint8(0))
	return instance
}

func (instance *fbftCore) receiveRequest(request *messages.Request) {
	nodes := instance.nodes.Load().(*nodesInfo)
	isMaster := nodes.local == nodes.master
	if !isMaster {
		log.Warn("only master process request.")
		return
	}
	log.Info("stop timeout master with view num %d.", instance.viewChange.GetCurrentViewNum())
	instance.stopChangeViewTimer()
	signature := request.Payload.Header.SigData
	if 0 == len(signature) {
		log.Error("request must have signature from producer.")
		return
	}
	if request.Account.Address != nodes.local.Address {
		log.Info("request from %x, not from %x.", request.Account.Address, nodes.master.Address)
		_, err := utils.VerifyPayload(request.Payload, instance.enableLocalVerifySignature)
		if nil != err {
			log.Error("proposal verified failed with error %v.", err)
			return
		}
	}
	content := instance.consensusPlugin.Add(request.Payload.Header.MixDigest, request.Payload)
	signData, err := utils.SignPayload(nodes.local, request.Payload.Header.MixDigest)
	if nil != err {
		log.Error("archive proposal signature failed with error %v.", err)
		return
	}
	if !content.AddSignature(nodes.local, signData) {
		log.Error("add signature to digest %v by account %d failed.",
			request.Payload.Header.MixDigest, nodes.local.Extension.Id)
		return
	}
	err = content.SetState(common.InConsensus)
	if nil != err {
		log.Error("set content state of %v failed with %v.", request.Payload.Header.MixDigest, err)
		return
	}
	proposal := messages.Message{
		MessageType: messages.ProposalMessageType,
		PayLoad: &messages.ProposalMessage{
			Proposal: &messages.Proposal{
				Account:   nodes.local,
				Timestamp: request.Timestamp,
				Payload:   request.Payload,
				Signature: signData,
			},
		},
	}
	rawData, err := messages.EncodeMessage(proposal)
	if nil != err {
		log.Error("marshal proposal msg failed with %v.", err)
		return
	}
	messages.BroadcastPeersFilter(rawData, proposal.MessageType, request.Payload.Header.MixDigest, nodes.peers, nodes.local)
	go instance.waitResponse(request.Payload.Header.MixDigest)
}

func (instance *fbftCore) waitResponse(digest types.Hash) {
	timeToCollectResponseMsg := time.NewTimer(time.Duration(instance.coreTimer.timeToCollectResponseMsg) * time.Millisecond)
	content, err := instance.consensusPlugin.GetContentByHash(digest)
	if nil != err {
		log.Error("get content of %v failed with %v.", digest, err)
		return
	}
	for {
		select {
		case <-timeToCollectResponseMsg.C:
			log.Warn("try to collect responses has to be overtime.")
			signatures, err := instance.maybeCommit(digest)
			if nil == err {
				content.SetState(common.ToConsensus)
				if instance.enableEmptyBlock || 0 != len(content.GetContentPayload().(*types.Block).Transactions) {
					instance.consensusPlugin.SetLatestBlockHeight(content.GetContentPayload().(*types.Block).Header.Height)
				}
			}
			consensusResult := messages.ConsensusResult{
				Signatures: signatures,
				Result:     err,
			}
			instance.result <- consensusResult
			return
		case signal := <-instance.signal:
			log.Debug("receive signal of %v.", signal)
			signatures, err := instance.maybeCommit(digest)
			if nil == err {
				content.SetState(common.ToConsensus)
				if instance.enableEmptyBlock || 0 != len(content.GetContentPayload().(*types.Block).Transactions) {
					instance.consensusPlugin.SetLatestBlockHeight(content.GetContentPayload().(*types.Block).Header.Height)
				}
				consensusResult := messages.ConsensusResult{
					Signatures: signatures,
					Result:     err,
				}
				instance.result <- consensusResult
				timeToCollectResponseMsg.Stop()
				log.Info("receive satisfied responses before overtime")
				return
			}
			log.Warn("get consensus result is error %v for current response.", err)
		}
	}
}

func (instance *fbftCore) receiveProposal(proposal *messages.Proposal) {
	proposalBlockHeight := proposal.Payload.Header.Height
	blockChain, err := repository.NewLatestStateRepository()
	if nil != err {
		panic(fmt.Errorf("get latest state block chain failed with err %v", err))
	}
	currentBlockHeight := blockChain.GetCurrentBlockHeight()
	if proposalBlockHeight <= currentBlockHeight {
		log.Warn("master is older than ourself, will ignore outdated proposal")
		return
	}

	nodes := instance.nodes.Load().(*nodesInfo)
	instance.stopChangeViewTimer()
	// if fall back, so sync block s first
	if proposalBlockHeight != common.DefaultBlockHeight && currentBlockHeight < proposalBlockHeight-1 {
		log.Warn("may be master info is wrong, which block height is %d, while received is %d, so change master to %d.",
			currentBlockHeight, proposalBlockHeight, proposal.Account.Extension.Id)
		instance.stopChangeViewTimer()

		newNodes := nodes.clone()
		newNodes.master = proposal.Account
		instance.nodes.Store(newNodes)

		blockSyncReq := &blockSyncRequest{
			start:  currentBlockHeight + 1,
			end:    proposalBlockHeight,
			target: proposal.Account,
		}
		log.Info("block height (%d) of the [node-%d] fall behind with the height (%d) of [node-%d]. will sync block from that node", currentBlockHeight,
			newNodes.local.Extension.Id, proposalBlockHeight-1, proposal.Account.Extension.Id)
		pushOrReplace(instance.blockSyncChan, blockSyncReq)
		return
	}
	// TODO: add view num to determine thr right master
	if nodes.master != proposal.Account {
		log.Error("proposal must from master %d, while it from %d in fact.",
			nodes.master.Extension.Id, proposal.Account.Extension.Id)
		return
	}
	// after sync, if master still not in inconsistent, try to change view
	isMaster := nodes.local == nodes.master
	if isMaster {
		log.Error("master will not receive proposal form itself %d.", proposal.Account.Extension.Id)
		return
	}
	go instance.waitMasterTimeout(time.Duration(instance.coreTimer.timeToWaitCommitMsg) * time.Millisecond)
	if !utils.SignatureVerify(nodes.master, proposal.Signature, proposal.Payload.Header.MixDigest) {
		log.Error("proposal signature not from master, please confirm.")
		return
	}
	instance.consensusPlugin.Add(proposal.Payload.Header.MixDigest, proposal.Payload)
	_, err = utils.VerifyPayload(proposal.Payload, instance.enableLocalVerifySignature)
	if nil != err {
		log.Error("proposal verified failed with error %v.", err)
		return
	}
	signData, err := utils.SignPayload(nodes.local, proposal.Payload.Header.MixDigest)
	if nil != err {
		log.Error("archive proposal signature failed with error %v.", err)
		return
	}
	response := messages.Message{
		MessageType: messages.ResponseMessageType,
		PayLoad: &messages.ResponseMessage{
			Response: &messages.Response{
				Account:     nodes.local,
				Timestamp:   proposal.Timestamp,
				Digest:      proposal.Payload.Header.MixDigest,
				Signature:   signData,
				BlockHeight: proposal.Payload.Header.Height,
			},
		},
	}
	msgRaw, err := messages.EncodeMessage(response)
	if nil != err {
		log.Error("encode proposal msg failed with %v.", err)
		return
	}
	messages.Unicast(nodes.master, msgRaw, response.MessageType, response.PayLoad.(*messages.ResponseMessage).Response.Digest)
}

func (instance *fbftCore) sendSyncBlockMsg(start uint64, end uint64, target account.Account) {
	nodes := instance.nodes.Load().(*nodesInfo)
	// TODO: sync blocks per block
	for index := start; index <= end; index++ {
		syncBlockRequest := messages.Message{
			MessageType: messages.SyncBlockReqMessageType,
			PayLoad: &messages.SyncBlockReqMessage{
				SyncBlockReq: &messages.SyncBlockReq{
					Account:    nodes.local,
					Timestamp:  time.Now().Unix(),
					BlockStart: index,
					BlockEnd:   index,
				},
			},
		}
		msgRaw, err := messages.EncodeMessage(syncBlockRequest)
		if nil != err {
			log.Error("encode syncBlockRequest msg failed with %v.", err)
			return
		}
		messages.Unicast(target, msgRaw, syncBlockRequest.MessageType, types.Hash{})
	}
}

func (instance *fbftCore) receiveSyncBlockRequest(request *messages.SyncBlockReq) {
	chain, err := repository.NewLatestStateRepository()
	if nil != err {
		panic("get new latest block state block chain failed.")
	}
	syncBlocks := make([]*types.Block, 0)
	for index := request.BlockStart; index <= request.BlockEnd; index++ {
		block, err := chain.GetBlockByHeight(index)
		if nil != err {
			log.Error("get block by height failed with err %v.", err)
			continue
		}
		syncBlocks = append(syncBlocks, block)
	}
	syncBlockResponse := messages.Message{
		MessageType: messages.SyncBlockRespMessageType,
		PayLoad: &messages.SyncBlockRespMessage{
			SyncBlockResp: &messages.SyncBlockResp{
				Blocks: syncBlocks,
			},
		},
	}
	msgRaw, err := messages.EncodeMessage(syncBlockResponse)
	if nil != err {
		log.Error("encode syncBlockResponse msg failed with %v.", err)
		return
	}
	messages.Unicast(request.Account, msgRaw, syncBlockResponse.MessageType, types.Hash{})
}

func (instance *fbftCore) receiveSyncBlockResponse(response *messages.SyncBlockResp) {
	blocks := response.Blocks
	log.Debug("receive sync %d block response.", len(blocks))
	for _, block := range blocks {
		chain, err := repository.NewLatestStateRepository()
		if nil != err {
			panic("new latest state block chain failed.")
		}
		currentBlockHeight := chain.GetCurrentBlockHeight()
		if currentBlockHeight < block.Header.Height {
			receipt, err := utils.VerifyPayloadUseExistedRepository(chain, block, instance.enableSyncVerifySignature)
			if nil != err {
				log.Error("verify failed with err %v.", err)
				continue
			}
			err = chain.WriteBlockWithReceipts(block, receipt)
			if nil != err {
				log.Error("write block with receipts failed with error %v.", err)
				continue
			}
		}
	}
}

func (instance *fbftCore) maybeCommit(digest types.Hash) ([][]byte, error) {
	nodes := instance.nodes.Load().(*nodesInfo)
	content, err := instance.consensusPlugin.GetContentByHash(digest)
	if nil != err {
		log.Error("get content of %v failed with %v.", digest, err)
		return make([][]byte, 0), fmt.Errorf("get content of %v failed with %v", digest, err)
	}
	signatures := content.Signatures()
	if uint8(len(signatures)) < uint8(len(nodes.peers))-instance.tolerance.Load().(uint8) {
		log.Warn("signature not satisfied which need %d, while receive %d now",
			uint8(len(nodes.peers))-instance.tolerance.Load().(uint8), len(signatures))
		return signatures, fmt.Errorf("signature not satisfy")
	}
	return signatures, nil
}

func (instance *fbftCore) receiveResponse(response *messages.Response) {
	nodes := instance.nodes.Load().(*nodesInfo)
	currentBlockHeight := instance.consensusPlugin.GetLatestBlockHeight()
	if response.BlockHeight <= currentBlockHeight {
		log.Warn("the response %d from node %d exceed deadline which is %d.",
			response.BlockHeight, response.Account.Extension.Id, currentBlockHeight)
		return
	}
	content, err := instance.consensusPlugin.GetContentByHash(response.Digest)
	if nil != err {
		log.Error("get content of %v from response failed with %v.", response.Digest, err)
		return
	}
	if common.ToConsensus != content.State() {
		isMaster := nodes.local == nodes.master
		if !isMaster {
			log.Info("only master need to process response.")
			return
		}
		if !utils.SignatureVerify(response.Account, response.Signature, response.Digest) {
			log.Error("signature and response sender not in coincidence.")
			return
		}
		if content.AddSignature(response.Account, response.Signature) {
			log.Debug("add commit response message from node %d.", response.Account.Extension.Id)
		} else {
			existSign, _ := content.GetSignByAccount(response.Account)
			if !bytes.Equal(existSign[:], response.Signature[:]) {
				log.Warn("receive diff signature from same validator %x, which exists is %x, while received is %x.",
					response.Account.Address, existSign, response.Signature)
			}
		}
		instance.signal <- common.ReceiveResponseSignal
	} else {
		log.Warn("consensus content state has reached %d, so ignore response from node %d.",
			common.ToConsensus, response.Account.Extension.Id)
	}
}

func (instance *fbftCore) tryToCommit(block *types.Block, result bool) {
	nodes := instance.nodes.Load().(*nodesInfo)
	commit := &messages.Commit{
		Account:    nodes.local,
		Timestamp:  time.Now().Unix(),
		Digest:     block.Header.MixDigest,
		Signatures: block.Header.SigData,
		BlockHash:  block.HeaderHash,
		Result:     result,
	}
	instance.sendCommit(commit, block)
}

func (instance *fbftCore) sendCommit(commit *messages.Commit, block *types.Block) {
	nodes := instance.nodes.Load().(*nodesInfo)
	committed := messages.Message{
		MessageType: messages.CommitMessageType,
		PayLoad: &messages.CommitMessage{
			Commit: commit,
		},
	}
	msgRaw, err := messages.EncodeMessage(committed)
	if nil != err {
		log.Error("EncodeMessage failed with %v.", err)
		return
	}
	if !commit.Result {
		log.Error("send the consensus with failed result.")
		messages.BroadcastPeersFilter(msgRaw, committed.MessageType, commit.Digest, nodes.peers, nodes.local)
		instance.eventCenter.Notify(types.EventConsensusFailed, nil)
	} else {
		log.Info("receive the successful consensus")
		messages.BroadcastPeersFilter(msgRaw, committed.MessageType, commit.Digest, nodes.peers, nodes.local)
		instance.commitBlock(block)
	}
}

func (instance *fbftCore) receiveCommit(commit *messages.Commit) {
	log.Debug("stop timeout master with view num %d.", instance.viewChange.GetCurrentViewNum())
	instance.stopChangeViewTimer()
	if !commit.Result {
		log.Error("receive commit is consensus error.")
		nodes := instance.nodes.Load().(*nodesInfo)
		newNodes := nodes.clone()
		newNodes.master = commit.Account
		instance.nodes.Store(newNodes)
		instance.eventCenter.Notify(types.EventConsensusFailed, nil)
		return
	}
	content, err := instance.consensusPlugin.GetContentByHash(commit.Digest)
	if nil != err {
		log.Error("get content of %v from commit failed with %v.", commit.Digest, err)
		return
	}
	payload := content.GetContentPayload()
	// TODO: verify signature loop
	payload.(*types.Block).Header.SigData = commit.Signatures
	payload.(*types.Block).HeaderHash = common.HeaderHash(payload.(*types.Block))
	if !(payload.(*types.Block).HeaderHash == commit.BlockHash) {
		log.Error("receive commit not consist, commit is %x, while compute is %x.",
			commit.BlockHash, payload.(*types.Block).HeaderHash)
		return
	}
	instance.commitBlock(payload.(*types.Block))
}

func (instance *fbftCore) commitBlock(block *types.Block) {
	instance.consensusPlugin.Remove(block.Header.MixDigest)
	if !instance.enableEmptyBlock && 0 == len(block.Transactions) {
		log.Warn("block without transaction.")
		instance.eventCenter.Notify(types.EventBlockWithoutTxs, nil)
		return
	}
	instance.blockSwitch <- block
	log.Info("try to commit block %d with hash %x to block switch.", block.Header.Height, block.HeaderHash)
}

func (instance *fbftCore) receiveChangeViewReq(viewChangeReq *messages.ViewChangeReq) {
	var nodes []account.Account
	currentViewNum := instance.viewChange.GetCurrentViewNum()
	if viewChangeReq.ViewNum <= currentViewNum {
		log.Warn("current viewNum %d no less than received %d, so ignore it.", currentViewNum, viewChangeReq.ViewNum)
		return
	}
	nodeInfos := instance.nodes.Load().(*nodesInfo)
	viewRequests, err := instance.viewChange.AddViewRequest(viewChangeReq.ViewNum, uint8(len(nodeInfos.peers))-instance.tolerance.Load().(uint8))
	if nil != err {
		log.Error("Add view request failed with error %v.", err)
		return
	}
	viewRequestState := viewRequests.GetViewRequestState()
	if viewRequestState != common.ViewEnd {
		// verify view change request signature
		for _, node := range viewChangeReq.Nodes {
			viewRequestState = viewRequests.ReceiveViewRequestByAccount(node)
		}
		// viewRequestState = viewRequests.ReceiveViewRequestByAccount(instance.nodes.local)
	}
	if viewRequestState == common.ViewEnd {
		instance.sendChangeViewReq(nodes, viewChangeReq.ViewNum)
		instance.stopChangeViewTimer()
		nodes = viewRequests.GetReceivedAccounts()
		instance.viewChange.SetCurrentViewNum(viewChangeReq.ViewNum)

		newNodesInfo := nodeInfos.clone()
		newNodesInfo.master = utils.GetAccountWithMinId(nodes)
		instance.nodes.Store(newNodesInfo)
		instance.eventCenter.Notify(types.EventMasterChange, nil)
		log.Info("now reach to consensus for viewNum %d and new master is %d.",
			viewChangeReq.ViewNum, newNodesInfo.master.Extension.Id)
	}
}

func (instance *fbftCore) sendChangeViewReq(nodes []account.Account, newView uint64) {
	nodeInfos := instance.nodes.Load().(*nodesInfo)
	syncBlockResMsg := messages.Message{
		MessageType: messages.ViewChangeMessageReqType,
		PayLoad: &messages.ViewChangeReqMessage{
			ViewChange: &messages.ViewChangeReq{
				Account:   nodeInfos.local,
				Nodes:     nodes,
				Timestamp: time.Now().Unix(),
				ViewNum:   newView,
			},
		},
	}
	msgRaw, err := messages.EncodeMessage(syncBlockResMsg)
	if nil != err {
		panic(fmt.Sprintf("marshal syncBlockResMsg msg failed with %v.", err))
	}
	// TODO: sign the digest
	messages.BroadcastPeersFilter(msgRaw, syncBlockResMsg.MessageType, types.Hash{}, nodeInfos.peers, nodeInfos.local)
}

func (instance *fbftCore) waitMasterTimeout(d time.Duration) {
	if !atomic.CompareAndSwapInt32(&instance.coreTimer.isRunning, 0, 1) {
		log.Info("previous timer is running, wait for previous finished")
		return
	}
	defer func() {
		atomic.StoreInt32(&instance.coreTimer.isRunning, 0)
	}()

	nodeInfos := instance.nodes.Load().(*nodesInfo)
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		currentViewNum := instance.viewChange.GetCurrentViewNum()
		requestViewNum := currentViewNum + 1
		log.Warn("master timeout, issue change view from %d to %d.", currentViewNum, requestViewNum)
		viewChangeReqMsg := messages.Message{
			MessageType: messages.ViewChangeMessageReqType,
			PayLoad: &messages.ViewChangeReqMessage{
				ViewChange: &messages.ViewChangeReq{
					Account:   nodeInfos.local,
					Nodes:     []account.Account{nodeInfos.local},
					Timestamp: time.Now().Unix(),
					ViewNum:   requestViewNum,
				},
			},
		}
		msgRaw, err := messages.EncodeMessage(viewChangeReqMsg)
		if nil != err {
			log.Error("marshal proposal msg failed with %v.", err)
			return
		}
		messages.BroadcastPeers(msgRaw, viewChangeReqMsg.MessageType, types.Hash{}, nodeInfos.peers)
		return
	case <-instance.coreTimer.timeToChangeViewStopChan:
		log.Info("stop change view timer")
		return
	}
}

// stop the instance change view timer
func (instance *fbftCore) stopChangeViewTimer() {
	// send stop signal to stop channel
	select {
	case instance.coreTimer.timeToChangeViewStopChan <- struct{}{}:
	default:
	}
}

func (instance *fbftCore) sendOnlineRequest() {
	nodeInfos := instance.nodes.Load().(*nodesInfo)
	log.Info("send online request.")
	chain, err := repository.NewLatestStateRepository()
	if nil != err {
		panic(fmt.Errorf("get latest state block chain failed with err %v", err))
	}
	currentBlockHeight := chain.GetCurrentBlockHeight()
	onlineMessage := messages.Message{
		MessageType: messages.OnlineRequestType,
		PayLoad: &messages.OnlineRequestMessage{
			OnlineRequest: &messages.OnlineRequest{
				Account:     nodeInfos.local,
				BlockHeight: currentBlockHeight,
				Timestamp:   time.Now().Unix(),
			},
		},
	}
	msgRaw, err := messages.EncodeMessage(onlineMessage)
	if nil != err {
		log.Error("marshal online request msg failed with %v.", err)
		return
	}
	// TODO: only need (instance.tolerance.Load().(uint8) + 1) agreement
	walterLevel := len(nodeInfos.peers) - int(instance.tolerance.Load().(uint8))
	currentHeight := instance.onlineWizard.GetCurrentHeight()
	var state common.OnlineState
	if currentBlockHeight == common.DefaultBlockHeight {
		_, state = instance.onlineWizard.AddOnlineResponse(
			currentBlockHeight, []account.Account{nodeInfos.local}, walterLevel, nodeInfos.master, instance.viewChange.GetCurrentViewNum())
		if currentHeight > currentBlockHeight {
			state = instance.onlineWizard.GetCurrentState()
		}
		if common.Online == state {
			instance.eventCenter.Notify(types.EventOnline, nil)
		}
	}
	messages.BroadcastPeersFilter(msgRaw, onlineMessage.MessageType, types.Hash{}, nodeInfos.peers, nodeInfos.local)
	return
}

func (instance *fbftCore) receiveOnlineRequest(request *messages.OnlineRequest) {
	nodeInfos := instance.nodes.Load().(*nodesInfo)
	chain, err := repository.NewLatestStateRepository()
	if nil != err {
		panic(fmt.Errorf("get latest state block chain failed with err %v", err))
	}
	currentBlockHeight := chain.GetCurrentBlockHeight()
	// sync block from remote
	if currentBlockHeight < request.BlockHeight && nodeInfos.local.Address != request.Account.Address {
		blockSyncReq := &blockSyncRequest{
			start:  currentBlockHeight + 1,
			end:    request.BlockHeight,
			target: request.Account,
		}
		pushOrReplace(instance.blockSyncChan, blockSyncReq)
	}

	currentViewNum := instance.viewChange.GetCurrentViewNum()
	log.Info("receive online request from node %d with height %d and local height is %d and local viewNum is %d.",
		request.Account.Extension.Id, request.BlockHeight, currentBlockHeight, currentViewNum)
	onlineResponse := messages.Message{
		MessageType: messages.OnlineResponseType,
		PayLoad: &messages.OnlineResponseMessage{
			OnlineResponse: &messages.OnlineResponse{
				Account:     nodeInfos.local,
				BlockHeight: currentBlockHeight,
				Nodes:       []account.Account{nodeInfos.local},
				Master:      nodeInfos.master,
				ViewNum:     currentViewNum,
				Timestamp:   time.Now().Unix(),
			},
		},
	}
	if common.DefaultBlockHeight == currentBlockHeight {
		state := instance.onlineWizard.GetCurrentStateByHeight(currentBlockHeight)
		if common.Online == state {
			log.Info("now has to be end of online and master is %d.", nodeInfos.master.Extension.Id)
		} else {
			walterLevel := len(nodeInfos.peers) - int(instance.tolerance.Load().(uint8))
			accounts := []account.Account{nodeInfos.local}
			if currentBlockHeight == request.BlockHeight {
				log.Info("init online, so add the received node of %d.", request.Account.Extension.Id)
				accounts = append(accounts, request.Account)
			}
			nodes, state := instance.onlineWizard.AddOnlineResponse(currentBlockHeight, accounts, walterLevel, nodeInfos.master, currentViewNum)
			if common.Online == state {
				log.Info("now has to be end of online and master is %d.", nodeInfos.master.Extension.Id)
				instance.eventCenter.Notify(types.EventOnline, nil)
			}
			log.Info("now receive %d response for block height %d and state is %v.", len(nodes), currentBlockHeight, state)
			onlineResponse.PayLoad.(*messages.OnlineResponseMessage).OnlineResponse.Nodes = nodes
			msgRaw, err := messages.EncodeMessage(onlineResponse)
			if nil != err {
				log.Error("marshal online request msg failed with %v.", err)
				panic("marshal online request msg failed.")
			}
			// init online, so broadcast it
			messages.BroadcastPeersFilter(msgRaw, onlineResponse.MessageType, types.Hash{}, nodeInfos.peers, nodeInfos.local)
			return
		}
	}
	msgRaw, err := messages.EncodeMessage(onlineResponse)
	if nil != err {
		log.Error("marshal online response msg failed with %v.", err)
		return
	}
	// not online first time, so just return result
	messages.Unicast(request.Account, msgRaw, onlineResponse.MessageType, types.Hash{})
}

func (instance *fbftCore) receiveOnlineResponse(response *messages.OnlineResponse) {
	nodeInfos := instance.nodes.Load().(*nodesInfo)
	log.Info("receive online response from node %d with height %d and viewNum %d and master %d.",
		response.Account.Extension.Id, response.BlockHeight, response.ViewNum, response.Master.Extension.Id)
	chain, err := repository.NewLatestStateRepository()
	if nil != err {
		panic(fmt.Errorf("get latest state block chain failed with err %v", err))
	}
	currentBlockHeight := chain.GetCurrentBlockHeight()
	if response.BlockHeight < currentBlockHeight {
		// sync block from remote
		blockSyncReq := &blockSyncRequest{
			start:  currentBlockHeight + 1,
			end:    response.BlockHeight,
			target: response.Account,
		}
		pushOrReplace(instance.blockSyncChan, blockSyncReq)
		log.Warn("block height received %d less than local %d, so ignore it.", response.BlockHeight, currentBlockHeight)
		return
	}
	state := instance.onlineWizard.GetResponseNodesStateByBlockHeight(response.BlockHeight)
	if common.Online == state {
		log.Warn("online state has to be online, so ignore response from node %d with height %d and master %d.",
			response.Account.Extension.Id, response.BlockHeight, response.Master.Extension.Id)
		return
	}
	walterLevel := len(nodeInfos.peers) - int(instance.tolerance.Load().(uint8))
	nodes, state := instance.onlineWizard.AddOnlineResponse(response.BlockHeight, response.Nodes, walterLevel, response.Master, response.ViewNum)
	if common.Online == state {
		log.Info("now online come to agree and master is %d and view num is %d.",
			response.Master.Extension.Id, response.ViewNum)
		instance.viewChange.SetCurrentViewNum(response.ViewNum)
		nodeInfos.master = response.Master
		instance.eventCenter.Notify(types.EventOnline, nil)
	}
	if common.DefaultViewNum == response.BlockHeight {
		log.Info("receive response when init online, so broadcast it.")
		onlineResponse := messages.Message{
			MessageType: messages.OnlineResponseType,
			PayLoad: &messages.OnlineResponseMessage{
				OnlineResponse: &messages.OnlineResponse{
					Account:     nodeInfos.local,
					Nodes:       nodes,
					BlockHeight: response.BlockHeight,
					Master:      response.Master,
					Timestamp:   time.Now().Unix(),
				},
			},
		}
		msgRaw, err := messages.EncodeMessage(onlineResponse)
		if nil != err {
			log.Error("marshal online response msg failed with %v.", err)
			return
		}
		messages.BroadcastPeersFilter(msgRaw, onlineResponse.MessageType, types.Hash{}, nodeInfos.peers, nodeInfos.local)
	}
}

func (instance *fbftCore) ProcessEvent(e utils.Event) utils.Event {
	nodeInfos := instance.nodes.Load().(*nodesInfo)
	var err error
	switch et := e.(type) {
	case *messages.Request:
		log.Info("receive request from replica %d with digest %x.",
			nodeInfos.local.Extension.Id, et.Payload.Header.MixDigest)
		instance.receiveRequest(et)
	case *messages.Proposal:
		log.Info("receive proposal from replica %d with digest %x.",
			et.Account.Extension.Id, et.Payload.Header.MixDigest)
		instance.receiveProposal(et)
	case *messages.Response:
		log.Info("receive response from replica %d with digest %x.",
			et.Account.Extension.Id, et.Digest)
		instance.receiveResponse(et)
	case *messages.Commit:
		log.Info("receive commit from replica %d with digest %x.",
			et.Account.Extension.Id, et.Digest)
		instance.receiveCommit(et)
	case *messages.SyncBlockReq:
		log.Info("receive sync block request from replica %d with start %d and end %d.",
			et.Account.Extension.Id, et.BlockStart, et.BlockEnd)
		instance.receiveSyncBlockRequest(et)
	case *messages.SyncBlockResp:
		log.Info("receive sync block response of %d blocks.", len(et.Blocks))
		instance.receiveSyncBlockResponse(et)
	case *messages.ViewChangeReq:
		log.Info("receive view change request from node %d and viewNum %d.",
			et.Account.Extension.Id, et.ViewNum)
		instance.receiveChangeViewReq(et)
	case *messages.OnlineRequest:
		log.Info("receive online request from node %d and block height %d.",
			et.Account.Extension.Id, et.BlockHeight)
		instance.receiveOnlineRequest(et)
	case *messages.OnlineResponse:
		log.Info("receive online response from node %d and block height %d with master %d.",
			et.Account.Extension.Id, et.BlockHeight, et.Master.Extension.Id)
		instance.receiveOnlineResponse(et)
	default:
		log.Warn("replica %d received an unknown message type %v",
			nodeInfos.local.Extension.Id, et)
		err = fmt.Errorf("not support type %v", et)
	}
	return err
}

func (instance *fbftCore) Start() {
	nodeInfos := instance.nodes.Load().(*nodesInfo)
	url := nodeInfos.local.Extension.Url
	localAddress, _ := net.ResolveTCPAddr("tcp4", url)
	var tcpListener, err = net.Listen("tcp", "0.0.0.0"+":"+strconv.Itoa(localAddress.Port))
	if err != nil {
		panic(fmt.Errorf("failed to start listener, as:%v", err))
	}
	defer func() {
		tcpListener.Close()
	}()
	log.Info("service start and waiting to be connected ...")
	go instance.blockSyncHandler()
	for {
		conn, err := tcpListener.Accept()
		log.Debug("accept a new connection from %s", conn.RemoteAddr().String())
		if err != nil {
			continue
		}
		go handleClient(conn, instance)
	}
}

func handleClient(conn net.Conn, bft *fbftCore) {
	nodeInfos := bft.nodes.Load().(*nodesInfo)
	conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(connReadTimeOut)))
	reader := bufio.NewReaderSize(conn, common.MaxBufferLen)
	msg, err := messages.ReadMessage(reader)
	if nil != err {
		conn.Close()
		log.Error("read message failed with error %v.", err)
		return
	}
	conn.Close()
	payload := msg.PayLoad
	log.Debug("receive [%d] message from from %s", msg.MessageType, conn.RemoteAddr().String())
	switch msg.MessageType {
	case messages.RequestMessageType:
		log.Info("receive request message from producer")
		// TODO: separate producer and master, so client need send request to master
		request := payload.(*messages.RequestMessage).Request
		utils.SendEvent(bft, request)
	case messages.ProposalMessageType:
		proposal := payload.(*messages.ProposalMessage).Proposal
		log.Info("receive proposal message form node %d with payload %x.",
			proposal.Account.Extension.Id, proposal.Payload.Header.MixDigest)
		utils.SendEvent(bft, proposal)
	case messages.ResponseMessageType:
		response := payload.(*messages.ResponseMessage).Response
		log.Info("receive response message from node %d with payload %x.",
			response.Account.Extension.Id, response.Digest)
		if response.Account.Extension.Id == nodeInfos.master.Extension.Id {
			log.Warn("master will not receive response message from itself.")
			return
		}
		utils.SendEvent(bft, response)
	case messages.SyncBlockReqMessageType:
		syncBlock := payload.(*messages.SyncBlockReqMessage).SyncBlockReq
		log.Info("receive sync block request message from node %d", syncBlock.Account.Extension.Id)
		utils.SendEvent(bft, syncBlock)
	case messages.SyncBlockRespMessageType:
		syncBlock := payload.(*messages.SyncBlockRespMessage).SyncBlockResp
		log.Info("receive sync blocks response from master.")
		utils.SendEvent(bft, syncBlock)
	case messages.CommitMessageType:
		commit := payload.(*messages.CommitMessage).Commit
		utils.SendEvent(bft, commit)
	case messages.ViewChangeMessageReqType:
		viewChange := payload.(*messages.ViewChangeReqMessage).ViewChange
		utils.SendEvent(bft, viewChange)
	case messages.OnlineRequestType:
		onlineRequest := payload.(*messages.OnlineRequestMessage).OnlineRequest
		utils.SendEvent(bft, onlineRequest)
	case messages.OnlineResponseType:
		onlineResponse := payload.(*messages.OnlineResponseMessage).OnlineResponse
		utils.SendEvent(bft, onlineResponse)
	default:
		if nil == payload {
			log.Warn("receive handshake, omit it %v.", payload)
		} else {
			log.Error("not support type for %v.", payload)
		}
	}
	return
}

// sync block from other nodes
func (instance *fbftCore) blockSyncHandler() {
	// obtain blockchain instance to used to get current block
	bchain, err := repository.NewLatestStateRepository()
	if err != nil {
		panic(fmt.Sprintf("failed to create blockchain as: %v", err))
	}
	// start loop to wait for sync request
	for {
		syncReq := <-instance.blockSyncChan
		instance.doSyncBlock(bchain, syncReq)
	}
}

func (instance *fbftCore) doSyncBlock(bchain *repository.Repository, syncReq *blockSyncRequest) {
	// subscribe block commit event
	syncSuccessChan := make(chan interface{})
	subscriber := instance.eventCenter.Subscribe(types.EventBlockCommitted, func(v interface{}) {
		select {
		case syncSuccessChan <- v:
		default:
			log.Info("block sync process is running, will ignore this block commit event and continue previous sync process")
		}
	})
	defer instance.eventCenter.UnSubscribe(types.EventBlockCommitted, subscriber)

	// sync timeout ticker
	syncTimeOut := time.NewTicker(5 * time.Second)

	for {
		currentBlock := bchain.GetCurrentBlock()
		if currentBlock == nil || currentBlock.Header.Height >= syncReq.end {
			return
		}
		currentHeight := currentBlock.Header.Height
		instance.sendSyncBlockMsg(currentHeight+1, currentHeight+1, syncReq.target)
		select {
		case <-syncSuccessChan:
		case <-syncTimeOut.C:
			return // encounter an error when sync block from this node. end up syncing
		}
	}
}

// push the blockReq into sync queue. if queue is full, replace the oldest sync request
func pushOrReplace(blockReqChan chan *blockSyncRequest, blockReq *blockSyncRequest) {
	for {
		//push the req into queue
		select {
		case blockReqChan <- blockReq:
			return
		default:
			log.Debug("block sync queue is full")
		}

		//drop the first req in sync queue
		select {
		case <-blockReqChan:
			log.Info("drop the oldest req in sync queue")
			continue
		default:
			return
		}
	}
}
