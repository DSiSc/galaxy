package fbft

import (
	"bytes"
	"encoding/json"
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

type fbftCore struct {
	local       account.Account
	master      account.Account
	peers       []account.Account
	signature   *tools.SignData
	tolerance   uint8
	commit      bool
	digest      types.Hash
	result      chan *messages.ConsensusResult
	tunnel      chan int
	validator   map[types.Hash]*payloadSets
	eventCenter types.EventCenter
	blockSwitch chan<- interface{}
}

type payloadSets struct {
	block    *types.Block
	receipts types.Receipts
}

func NewFBFTCore(local account.Account, result chan *messages.ConsensusResult, blockSwitch chan<- interface{}) *fbftCore {
	return &fbftCore{
		local: local,
		signature: &tools.SignData{
			Signatures: make([][]byte, 0),
			SignMap:    make(map[account.Account][]byte),
		},
		result:      result,
		tunnel:      make(chan int),
		validator:   make(map[types.Hash]*payloadSets),
		blockSwitch: blockSwitch,
	}
}

func (instance *fbftCore) receiveRequest(request *messages.Request) {
	isMaster := instance.local == instance.master
	if !isMaster {
		log.Info("only master process request.")
		return
	}
	signature := request.Payload.Header.SigData
	if 1 != len(signature) {
		log.Error("request must have signature from producer.")
		return
	}
	receipts, err := tools.VerifyPayload(request.Payload)
	if nil != err {
		log.Error("proposal verified failed with error %v.", err)
		return
	}
	signData, err := tools.SignPayload(instance.local, request.Payload.Header.MixDigest)
	if nil != err {
		log.Error("archive proposal signature failed with error %v.", err)
		return
	}
	if values, ok := instance.validator[request.Payload.Header.MixDigest]; !ok {
		log.Info("add record payload %x.", request.Payload.Header.MixDigest)
		instance.validator[request.Payload.Header.MixDigest] = &payloadSets{
			block:    request.Payload,
			receipts: receipts,
		}
	} else {
		values.receipts = receipts
	}
	proposal := &messages.Message{
		MessageType: messages.ProposalMessageType,
		Payload: &messages.ProposalMessage{
			Proposal: &messages.Proposal{
				Id:        instance.local.Extension.Id,
				Timestamp: request.Timestamp,
				Payload:   request.Payload,
				Signature: signData,
			},
		},
	}
	msgRaw, err := json.Marshal(proposal)
	if nil != err {
		log.Error("marshal proposal msg failed with %v.", err)
		return
	}
	instance.digest = request.Payload.Header.MixDigest
	instance.signature.AddSignature(instance.local, signData)
	// filter master
	peers := tools.AccountFilter([]account.Account{instance.local}, instance.peers)
	messages.BroadcastPeers(msgRaw, proposal.MessageType, instance.digest, peers)
	go instance.waitResponse()
}

func (instance *fbftCore) waitResponse() {
	log.Warn("set timer with 2 second.")
	timer := time.NewTimer(5 * time.Second)
	for {
		select {
		case <-timer.C:
			log.Info("collect response timeout.")
			signatures, err := instance.maybeCommit()
			instance.commit = true
			consensusResult := &messages.ConsensusResult{
				Signatures: signatures,
				Result:     err,
			}
			instance.result <- consensusResult
			return
		case <-instance.tunnel:
			log.Debug("receive tunnel")
			signatures, err := instance.maybeCommit()
			if nil == err {
				instance.commit = true
				consensusResult := messages.ConsensusResult{
					Signatures: signatures,
					Result:     err,
				}
				instance.result <- &consensusResult
				log.Info("receive satisfied responses before timeout")
				return
			}
			log.Warn("get consensus result is error %v.", err)
		}
	}
}

func (instance *fbftCore) receiveProposal(proposal *messages.Proposal) {
	isMaster := instance.local == instance.master
	if isMaster {
		log.Info("master not need to process proposal.")
		return
	}
	if instance.master.Extension.Id != proposal.Id {
		log.Error("proposal must from master %d, while it from %d in fact.", instance.master.Extension.Id, proposal.Id)
		return
	}
	if !signDataVerify(instance.master, proposal.Signature, proposal.Payload.Header.MixDigest) {
		log.Error("proposal signature not from master, please confirm.")
		return
	}
	receipts, err := tools.VerifyPayload(proposal.Payload)
	if nil != err {
		log.Error("proposal verified failed with error %v.", err)
		return
	}
	signData, err := tools.SignPayload(instance.local, proposal.Payload.Header.MixDigest)
	if nil != err {
		log.Error("archive proposal signature failed with error %v.", err)
		return
	}

	if values, ok := instance.validator[proposal.Payload.Header.MixDigest]; !ok {
		log.Info("add record payload %x.", proposal.Payload.Header.MixDigest)
		instance.validator[proposal.Payload.Header.MixDigest] = &payloadSets{
			block:    proposal.Payload,
			receipts: receipts,
		}
	} else {
		values.receipts = receipts
	}
	response := &messages.Message{
		MessageType: messages.ResponseMessageType,
		Payload: &messages.ResponseMessage{
			Response: &messages.Response{
				Account:   instance.local,
				Timestamp: proposal.Timestamp,
				Digest:    proposal.Payload.Header.MixDigest,
				Signature: signData,
			},
		},
	}
	msgRaw, err := json.Marshal(response)
	if nil != err {
		log.Error("marshal proposal msg failed with %v.", err)
		return
	}
	messages.Unicast(instance.master, msgRaw, messages.ResponseMessageType, proposal.Payload.Header.MixDigest)
}

func (instance *fbftCore) maybeCommit() ([][]byte, error) {
	var reallySignature = make([][]byte, 0)
	if len(instance.signature.Signatures) != len(instance.signature.SignMap) {
		log.Error("length of signData[%d] and signMap[%d] does not match.", len(instance.signature.Signatures), len(instance.signature.SignMap))
		return reallySignature, fmt.Errorf("signData and signMap does not match")
	}
	var suspiciousAccount = make([]account.Account, 0)
	for account, sign := range instance.signature.SignMap {
		if signDataVerify(account, sign, instance.digest) {
			reallySignature = append(reallySignature, sign)
			continue
		}
		suspiciousAccount = append(suspiciousAccount, account)
		log.Warn("signature %x by account %x is invalid", sign, account)
	}
	if uint8(len(reallySignature)) < uint8(len(instance.peers))-instance.tolerance {
		log.Warn("really signature %d less than need %d.",
			len(reallySignature), uint8(len(instance.peers))-instance.tolerance)
		return reallySignature, fmt.Errorf("signature not satisfy")
	}
	return reallySignature, nil
}

func signDataVerify(account account.Account, sign []byte, digest types.Hash) bool {
	address, err := signature.Verify(digest, sign)
	if nil != err {
		log.Error("verify sign %v failed with err %s which expect from %x", sign, err, account.Address)
	}
	return account.Address == address
}

func (instance *fbftCore) receiveResponse(response *messages.Response) {
	if !instance.commit {
		isMaster := instance.local == instance.master
		if !isMaster {
			log.Info("only master need to process response.")
			return
		}
		if !bytes.Equal(instance.digest[:], response.Digest[:]) {
			log.Error("received response digest %x not in coincidence with reserved %x.",
				instance.digest, response.Digest)
			return
		}
		from := tools.GetAccountById(instance.peers, response.Account.Extension.Id)
		if !signDataVerify(from, response.Signature, instance.digest) {
			log.Error("signature and response sender not in coincidence.")
			return
		}
		if sign, ok := instance.signature.SignMap[from]; !ok {
			instance.signature.AddSignature(from, response.Signature)
			instance.tunnel <- 1
			log.Info("response from %x has been committed yet.", response.Account.Address)
		} else {
			log.Warn("receive duplicate signature from the same validator, ignore it.")
			if !bytes.Equal(sign, response.Signature) {
				log.Warn("receive a different signature from the same validator %x, which exists is %x, while response is %x.",
					from.Address, sign, response.Signature)
			}
		}
	} else {
		log.Info("response has be committed, ignore response from %x.", response.Account.Address)
	}
}

func (instance *fbftCore) SendCommit(commit *messages.Commit, block *types.Block) {
	committed := &messages.Message{
		MessageType: messages.CommitMessageType,
		Payload: &messages.CommitMessage{
			Commit: commit,
		},
	}
	msgRaw, err := json.Marshal(committed)
	if nil != err {
		log.Error("marshal commit msg failed with %v.", err)
		return
	}
	peers := tools.AccountFilter([]account.Account{instance.local}, instance.peers)
	if !commit.Result {
		log.Error("send the failed consensus.")
		messages.BroadcastPeers(msgRaw, committed.MessageType, commit.Digest, peers)
		instance.eventCenter.Notify(types.EventConsensusFailed, nil)
	} else {
		log.Info("receive the successful consensus")
		messages.BroadcastPeers(msgRaw, committed.MessageType, commit.Digest, peers)
		instance.commitBlock(block)
	}
}

func (instance *fbftCore) receiveCommit(commit *messages.Commit) {
	log.Info("receive commit from node %x", commit.Account.Address)
	if !commit.Result {
		log.Error("receive commit is consensus error.")
		instance.eventCenter.Notify(types.EventConsensusFailed, nil)
		return
	}
	if payload, ok := instance.validator[commit.Digest]; ok {
		payload.block.Header.SigData = commit.Signatures
		blockHash := common.HeaderHash(payload.block)
		if !bytes.Equal(blockHash[:], commit.BlockHash[:]) {
			log.Error("receive commit not consist, commit is %x, while compute is %x.", commit.BlockHash, blockHash)
			payload.block.Header.SigData = make([][]byte, 0)
			return
		}
		// TODO: verify signature loop
		chain, err := blockchain.NewBlockChainByBlockHash(payload.block.Header.PrevBlockHash)
		if nil != err {
			payload.block.Header.SigData = make([][]byte, 0)
			log.Error("get NewBlockChainByHash by hash %x failed with error %s.", payload.block.Header.PrevBlockHash, err)
			return
		}
		payload.block.HeaderHash = common.HeaderHash(payload.block)
		log.Info("begin write block %d with hash %x.", payload.block.Header.Height, payload.block.HeaderHash)
		err = chain.WriteBlockWithReceipts(payload.block, payload.receipts)
		if nil != err {
			payload.block.Header.SigData = make([][]byte, 0)
			log.Error("call WriteBlockWithReceipts failed with", payload.block.Header.PrevBlockHash, err)
		}
		log.Info("end write block %d with hash %x with success.", payload.block.Header.Height, payload.block.HeaderHash)
		return
	}
	log.Error("payload with digest %x not found, please confirm.", commit.Digest)
}

func (instance *fbftCore) commitBlock(block *types.Block) {
	instance.blockSwitch <- block
	log.Info("write block %d with hash %x success.", block.Header.Height, block.HeaderHash)
}

func (instance *fbftCore) ProcessEvent(e tools.Event) tools.Event {
	var err error
	switch et := e.(type) {
	case *messages.Request:
		log.Info("receive request %x from replica %d.", et.Payload.Header.MixDigest, instance.local.Extension.Id)
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
	default:
		log.Warn("replica %d received an unknown message type %v", instance.local.Extension.Id, et)
		err = fmt.Errorf("not support type %v", et)
	}
	if err != nil {
		log.Warn(err.Error())
	}
	return err
}

func (instance *fbftCore) Start(account account.Account) {
	url := account.Extension.Url
	log.Info("start server of url: %s.", url)
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

func handleConnection(tcpListener *net.TCPListener, fbft *fbftCore) {
	buffer := make([]byte, 2048)
	for {
		var conn, _ = tcpListener.AcceptTCP()
		n, err := conn.Read(buffer)
		if err != nil {
			log.Error("error when read connector %x.", err)
			return
		}
		log.Info("received a messages.")
		var msg messages.Message
		err = json.Unmarshal(buffer[:n], &msg)
		payload := msg.Payload
		switch msg.MessageType {
		case messages.RequestMessageType:
			request := payload.(*messages.RequestMessage).Request
			log.Info("receive a request message")
			tools.SendEvent(fbft, request)
		case messages.ProposalMessageType:
			proposal := payload.(*messages.ProposalMessage).Proposal
			log.Info("receive a proposal message form node %d with payload %x.",
				proposal.Id, proposal.Payload.Header.MixDigest)
			if proposal.Id != fbft.master.Extension.Id {
				log.Warn("only master can issue a proposal.")
				continue
			}
			tools.SendEvent(fbft, proposal)
		case messages.ResponseMessageType:
			response := payload.(*messages.ResponseMessage).Response
			log.Info("receive response message from node %d with payload %x.",
				response.Account.Extension.Id, response.Digest)
			if response.Account.Extension.Id == fbft.master.Extension.Id {
				log.Warn("master will not receive response message from itself.")
				continue
			}
			tools.SendEvent(fbft, response)
		case messages.CommitMessageType:
			commit := payload.(*messages.CommitMessage).Commit
			tools.SendEvent(fbft, commit)
		default:
			if nil == payload {
				log.Info("receive handshake, omit it.")
			} else {
				log.Error("not support type for %v.", payload)
			}
			return
		}
	}
}

func handleClient(conn net.Conn, bft *fbftCore) {
	log.Info("receive messages form other node.")
	defer conn.Close()
	buffer := make([]byte, 20480)
	len, err := conn.Read(buffer)
	if err != nil {
		log.Error("error when read connector %v.", err)
		return
	}
	if len == 0 {
		log.Error("read data length is 0.")
		return
	}
	var msg messages.Message
	err = json.Unmarshal(buffer[:len], &msg)
	if err != nil {
		log.Error("unmarshal failed with error %v.", err)
		return
	}
	payload := msg.Payload
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
		if proposal.Id != bft.master.Extension.Id {
			log.Warn("only master can issue a proposal.")
			return
		}
		tools.SendEvent(bft, proposal)
	case messages.ResponseMessageType:
		response := payload.(*messages.ResponseMessage).Response
		log.Info("receive response message from node %d with payload %x.",
			response.Account.Extension.Id, response.Digest)
		if response.Account.Extension.Id == bft.master.Extension.Id {
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
	default:
		if nil == payload {
			log.Info("receive handshake, omit it %v.", payload)
		} else {
			log.Error("not support type for %v.", payload)
		}
	}
}
