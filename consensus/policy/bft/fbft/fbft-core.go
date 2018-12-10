package fbft

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/DSiSc/blockchain"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/policy/bft/messages"
	"github.com/DSiSc/galaxy/consensus/policy/bft/tools"
	"github.com/DSiSc/validator/tools/account"
	"github.com/DSiSc/validator/tools/signature"
	"net"
	"time"
)

type nodesInfo struct {
	local  account.Account
	master account.Account
	peers  []account.Account
}

type fbftCore struct {
	nodes           *nodesInfo
	tolerance       uint8
	status          common.ViewStatus
	timeoutTimer    *time.Timer
	result          chan messages.ConsensusResult
	signal          chan common.MessageSignal
	eventCenter     types.EventCenter
	blockSwitch     chan<- interface{}
	consensusPlugin *tools.ConsensusPlugin
	viewChange      *tools.ViewChange
}

func NewFBFTCore(local account.Account, blockSwitch chan<- interface{}) *fbftCore {
	return &fbftCore{
		nodes:           &nodesInfo{local: local},
		status:          common.ViewNormal,
		result:          make(chan messages.ConsensusResult),
		signal:          make(chan common.MessageSignal),
		blockSwitch:     blockSwitch,
		consensusPlugin: tools.NewConsensusPlugin(),
		viewChange:      tools.NewViewChange(),
	}
}

func (instance *fbftCore) receiveRequest(request *messages.Request) {
	isMaster := instance.nodes.local == instance.nodes.master
	if !isMaster {
		log.Warn("only master process request.")
		return
	}
	instance.timeoutTimer.Stop()
	signature := request.Payload.Header.SigData
	if 1 != len(signature) {
		log.Error("request must have signature from producer.")
		return
	}
	_, err := tools.VerifyPayload(request.Payload)
	if nil != err {
		log.Error("proposal verified failed with error %v.", err)
		return
	}
	content := instance.consensusPlugin.Add(request.Payload.Header.MixDigest, request.Payload)
	signData, err := tools.SignPayload(instance.nodes.local, request.Payload.Header.MixDigest)
	if nil != err {
		log.Error("archive proposal signature failed with error %v.", err)
		return
	}
	if !content.AddSignature(instance.nodes.local, signData) {
		log.Error("add signature to digest %v by account %d failed.",
			request.Payload.Header.MixDigest, instance.nodes.local)
		return
	}
	err = content.SetState(tools.InConsensus)
	if nil != err {
		log.Error("set content state of %v failed with %v.", request.Payload.Header.MixDigest, err)
		return
	}
	proposal := messages.Message{
		MessageType: messages.ProposalMessageType,
		PayLoad: &messages.ProposalMessage{
			Proposal: &messages.Proposal{
				Id:        instance.nodes.local.Extension.Id,
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
	messages.BroadcastPeersFilter(rawData, proposal.MessageType, request.Payload.Header.MixDigest, instance.nodes.peers, instance.nodes.local)
	go instance.waitResponse(request.Payload.Header.MixDigest)
}

func (instance *fbftCore) waitResponse(digest types.Hash) {
	timer := time.NewTimer(5 * time.Second)
	content, err := instance.consensusPlugin.GetContentByHash(digest)
	if nil != err {
		log.Error("get content of %v failed with %v.", digest, err)
		return
	}
	for {
		select {
		case <-timer.C:
			log.Info("response has overtime.")
			signatures, err := instance.maybeCommit(digest)
			content.SetState(tools.ToConsensus)
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
				content.SetState(tools.ToConsensus)
				consensusResult := messages.ConsensusResult{
					Signatures: signatures,
					Result:     err,
				}
				instance.result <- consensusResult
				log.Info("receive satisfied responses before overtime")
				return
			}
			log.Warn("get consensus result is error %v for current response.", err)
		}
	}
}

func (instance *fbftCore) receiveProposal(proposal *messages.Proposal) {
	isMaster := instance.nodes.local == instance.nodes.master
	if isMaster {
		log.Warn("master not need to process proposal.")
		return
	}
	instance.timeoutTimer.Reset(10 * time.Second)
	if instance.nodes.master.Extension.Id != proposal.Id {
		log.Error("proposal must from master %d, while it from %d in fact.", instance.nodes.master.Extension.Id, proposal.Id)
		return
	}
	if !signDataVerify(instance.nodes.master, proposal.Signature, proposal.Payload.Header.MixDigest) {
		log.Error("proposal signature not from master, please confirm.")
		return
	}
	instance.consensusPlugin.Add(proposal.Payload.Header.MixDigest, proposal.Payload)
	_, err := tools.VerifyPayload(proposal.Payload)
	if nil != err {
		log.Error("proposal verified failed with error %v.", err)
		return
	}
	signData, err := tools.SignPayload(instance.nodes.local, proposal.Payload.Header.MixDigest)
	if nil != err {
		log.Error("archive proposal signature failed with error %v.", err)
		return
	}
	response := messages.Message{
		MessageType: messages.ResponseMessageType,
		PayLoad: &messages.ResponseMessage{
			Response: &messages.Response{
				Account:   instance.nodes.local,
				Timestamp: proposal.Timestamp,
				Digest:    proposal.Payload.Header.MixDigest,
				Signature: signData,
			},
		},
	}
	msgRaw, err := messages.EncodeMessage(response)
	if nil != err {
		log.Error("encode proposal msg failed with %v.", err)
		return
	}
	messages.Unicast(instance.nodes.master, msgRaw, response.MessageType, response.PayLoad.(*messages.ResponseMessage).Response.Digest)
}

func (instance *fbftCore) maybeCommit(digest types.Hash) ([][]byte, error) {
	content, err := instance.consensusPlugin.GetContentByHash(digest)
	if nil != err {
		log.Error("get content of %v failed with %v.", digest, err)
		return make([][]byte, 0), fmt.Errorf("get content of %v failed with %v", digest, err)
	}
	signatures := content.Signatures()
	if uint8(len(signatures)) < uint8(len(instance.nodes.peers))-instance.tolerance {
		log.Warn("signature not satisfied which need %d, while receive %d now",
			uint8(len(instance.nodes.peers))-instance.tolerance, len(signatures))
		return signatures, fmt.Errorf("signature not satisfy")
	}
	return signatures, nil
}

func signDataVerify(account account.Account, sign []byte, digest types.Hash) bool {
	address, err := signature.Verify(digest, sign)
	if nil != err {
		log.Error("verify sign %v failed with err %s which expect from %x", sign, err, account.Address)
	}
	return account.Address == address
}

func (instance *fbftCore) receiveResponse(response *messages.Response) {
	content, err := instance.consensusPlugin.GetContentByHash(response.Digest)
	if nil != err {
		log.Error("get content of %v from response failed with %v.", response.Digest, err)
		return
	}
	if tools.ToConsensus != content.State() {
		isMaster := instance.nodes.local == instance.nodes.master
		if !isMaster {
			log.Info("only master need to process response.")
			return
		}
		if !signDataVerify(response.Account, response.Signature, response.Digest) {
			log.Error("signature and response sender not in coincidence.")
			return
		}
		if content.AddSignature(response.Account, response.Signature) {
			log.Debug("commit response message from node %d.", response.Account.Extension.Id)
		} else {
			existSign, _ := content.GetSignByAccount(response.Account)
			if !bytes.Equal(existSign[:], response.Signature[:]) {
				log.Warn("receive diff signature from same validator %x, which exists is %x, while received is %x.",
					response.Account.Address, existSign, response.Signature)
			}
		}
		instance.signal <- common.ReceiveResponseSignal
	} else {
		log.Warn("consensus content state has reached %d, so ignore response from %x.",
			tools.ToConsensus, response.Account.Address)
	}
}

func (instance *fbftCore) SendCommit(commit *messages.Commit, block *types.Block) {
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
	// peers := tools.AccountFilter([]account.Account{instance.local}, instance.peers)
	if !commit.Result {
		log.Error("send the failed consensus.")
		// messages.BroadcastPeers(msgRaw, committed.MessageType, commit.Digest, peers)
		messages.BroadcastPeersFilter(msgRaw, committed.MessageType, commit.Digest, instance.nodes.peers, instance.nodes.local)
		instance.eventCenter.Notify(types.EventConsensusFailed, nil)
	} else {
		log.Info("receive the successful consensus")
		messages.BroadcastPeersFilter(msgRaw, committed.MessageType, commit.Digest, instance.nodes.peers, instance.nodes.local)
		instance.commitBlock(block)
	}
}

func (instance *fbftCore) receiveCommit(commit *messages.Commit) {
	if !commit.Result {
		log.Error("receive commit is consensus error.")
		instance.eventCenter.Notify(types.EventConsensusFailed, nil)
		return
	}
	instance.timeoutTimer.Stop()
	content, err := instance.consensusPlugin.GetContentByHash(commit.Digest)
	if nil != err {
		log.Error("get content of %v from commit failed with %v.", commit.Digest, err)
		return
	}
	payload := content.GetContentPayloadByHash(commit.Digest)
	// TODO: verify signature loop
	payload.(*types.Block).Header.SigData = commit.Signatures
	payload.(*types.Block).HeaderHash = common.HeaderHash(payload.(*types.Block))
	if !bytes.Equal(payload.(*types.Block).HeaderHash[:], commit.BlockHash[:]) {
		log.Error("receive commit not consist, commit is %x, while compute is %x.",
			commit.BlockHash, payload.(*types.Block).HeaderHash)
		return
	}
	instance.commitBlock(payload.(*types.Block))
}

func (instance *fbftCore) commitBlock(block *types.Block) {
	// delete(instance.validator, block.Header.MixDigest)
	// instance.consensusPlugin.Remove(block.Header.MixDigest)
	chain, _ := blockchain.NewBlockChainByBlockHash(block.Header.PrevBlockHash)
	preBlock, _ := chain.GetBlockByHash(block.Header.PrevBlockHash)
	instance.consensusPlugin.Remove(preBlock.HeaderHash)
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
	viewRequests, err := instance.viewChange.AddViewRequest(viewChangeReq.ViewNum, uint8(len(instance.nodes.peers))-instance.tolerance)
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
		viewRequestState = viewRequests.ReceiveViewRequestByAccount(instance.nodes.local)
	}
	nodes = viewRequests.GetReceivedAccounts()
	if viewRequestState == common.ViewEnd {
		// come to consensus for new view number
		instance.viewChange.SetCurrentViewNum(viewChangeReq.ViewNum)
		instance.nodes.master = tools.GetNodeAccountWithMinId(nodes)
		instance.eventCenter.Notify(types.EventMasterChange, nil)
		log.Info("now reach to consensus for viewNum %d and new master is %vd.",
			viewChangeReq.ViewNum, instance.nodes.master.Extension.Id)
	}
	instance.sendChangeViewReq(nodes, viewChangeReq.ViewNum)
}

func (instance *fbftCore) sendChangeViewReq(nodes []account.Account, newView uint64) {
	syncBlockResMsg := messages.Message{
		MessageType: messages.ViewChangeMessageReqType,
		PayLoad: &messages.ViewChangeReqMessage{
			ViewChange: &messages.ViewChangeReq{
				Id:        instance.nodes.local.Extension.Id,
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
	messages.BroadcastPeersFilter(msgRaw, syncBlockResMsg.MessageType, types.Hash{}, instance.nodes.peers, instance.nodes.local)
}

func (self *fbftCore) waitMasterTimeout(timer *time.Timer) {
	for {
		select {
		case <-timer.C:
			currentViewNum := self.viewChange.GetCurrentViewNum()
			requestViewNum := currentViewNum + 1
			log.Warn("master timeout, issue change view from %d to %d.", currentViewNum, requestViewNum)
			viewChangeReqMsg := messages.Message{
				MessageType: messages.ViewChangeMessageReqType,
				PayLoad: &messages.ViewChangeReqMessage{
					ViewChange: &messages.ViewChangeReq{
						Nodes:     []account.Account{self.nodes.local},
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
			messages.BroadcastPeers(msgRaw, viewChangeReqMsg.MessageType, types.Hash{}, self.nodes.peers)
			return
		}
	}
}

func (instance *fbftCore) ProcessEvent(e tools.Event) tools.Event {
	var err error
	switch et := e.(type) {
	case *messages.Request:
		log.Info("receive request %x from replica %d.", et.Payload.Header.MixDigest, instance.nodes.local.Extension.Id)
		instance.receiveRequest(et)
	case *messages.Proposal:
		log.Info("receive proposal from replica %d with digest %x.", et.Id, et.Payload.Header.MixDigest)
		instance.receiveProposal(et)
	case *messages.Response:
		log.Info("receive response from replica %d with digest %x.", et.Account.Extension.Id, et.Digest)
		instance.receiveResponse(et)
	case *messages.Commit:
		log.Info("receive commit from replica %d with digest %x.", et.Account.Extension.Id, et.Digest)
		instance.receiveCommit(et)
	case *messages.ViewChangeReq:
		log.Info("receive view change request from node %d and viewNum %d.", et.Id, et.ViewNum)
		instance.receiveChangeViewReq(et)
	default:
		log.Warn("replica %d received an unknown message type %v", instance.nodes.local.Extension.Id, et)
		err = fmt.Errorf("not support type %v", et)
	}
	return err
}

func (instance *fbftCore) Start() {
	url := instance.nodes.local.Extension.Url
	localAddress, _ := net.ResolveTCPAddr("tcp4", url)
	var tcpListener, err = net.ListenTCP("tcp", localAddress)
	if err != nil {
		log.Error("listen error：%v.", err)
		return
	}
	defer func() {
		tcpListener.Close()
	}()
	log.Info("service start and waiting to be connected ...")
	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn, instance)
	}
}

func handleClient(conn net.Conn, bft *fbftCore) {
	defer conn.Close()
	reader := bufio.NewReaderSize(conn, common.MAX_BUF_LEN)
	msg, err := messages.ReadMessage(reader)
	if nil != err {
		log.Error("read message failed with error %v.", err)
		return
	}
	payload := msg.PayLoad
	switch msg.MessageType {
	case messages.RequestMessageType:
		log.Info("receive request message from producer")
		// TODO: separate producer and master, so client need send request to master
		request := payload.(*messages.RequestMessage).Request
		tools.SendEvent(bft, request)
	case messages.ProposalMessageType:
		proposal := payload.(*messages.ProposalMessage).Proposal
		log.Info("receive proposal message form node %d with payload %x.",
			proposal.Id, proposal.Payload.Header.MixDigest)
		if proposal.Id != bft.nodes.master.Extension.Id {
			log.Warn("only master can issue a proposal.")
			return
		}
		tools.SendEvent(bft, proposal)
	case messages.ResponseMessageType:
		response := payload.(*messages.ResponseMessage).Response
		log.Info("receive response message from node %d with payload %x.",
			response.Account.Extension.Id, response.Digest)
		if response.Account.Extension.Id == bft.nodes.master.Extension.Id {
			log.Warn("master will not receive response message from itself.")
			return
		}
		tools.SendEvent(bft, response)
	case messages.SyncBlockReqMessageType:
		syncBlock := payload.(*messages.SyncBlockReqMessage).SyncBlock
		log.Info("receive sync block message from node %d", syncBlock.Node.Extension.Id)
		tools.SendEvent(bft, syncBlock)
	case messages.SyncBlockRespMessageType:
		syncBlock := payload.(*messages.SyncBlockRespMessage).SyncBlock
		log.Info("receive sync blocks from master.")
		tools.SendEvent(bft, syncBlock)
	case messages.CommitMessageType:
		commit := payload.(*messages.CommitMessage).Commit
		tools.SendEvent(bft, commit)
	case messages.ViewChangeMessageReqType:
		viewChange := payload.(*messages.ViewChangeReqMessage).ViewChange
		tools.SendEvent(bft, viewChange)
	default:
		if nil == payload {
			log.Warn("receive handshake, omit it %v.", payload)
		} else {
			log.Error("not support type for %v.", payload)
		}
	}
}
