package bft

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/DSiSc/blockchain"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/messages"
	tools "github.com/DSiSc/galaxy/consensus/utils"
	"github.com/DSiSc/validator/tools/account"
	"github.com/DSiSc/validator/tools/signature"
	"github.com/DSiSc/validator/worker"
	"net"
	"sync"
	"time"
)

type bftCore struct {
	local       account.Account
	mutex       sync.RWMutex
	master      account.Account
	peers       []account.Account
	signature   *signData
	tolerance   uint8
	commit      bool
	digest      types.Hash
	result      chan *messages.ConsensusResult
	tunnel      chan int
	validator   map[types.Hash]*payloadSets
	payloads    map[types.Hash]*types.Block
	eventCenter types.EventCenter
}

type signData struct {
	signatures [][]byte
	signMap    map[account.Account][]byte
}

func (s *signData) addSignature(account account.Account, sign []byte) {
	log.Info("add %x signature.", account.Address)
	s.signMap[account] = sign
	s.signatures = append(s.signatures, sign)
}

type payloadSets struct {
	block    *types.Block
	receipts types.Receipts
}

func NewBFTCore(local account.Account, result chan *messages.ConsensusResult) *bftCore {
	return &bftCore{
		local: local,
		signature: &signData{
			signatures: make([][]byte, 0),
			signMap:    make(map[account.Account][]byte),
		},
		result:    result,
		tunnel:    make(chan int),
		validator: make(map[types.Hash]*payloadSets),
		payloads:  make(map[types.Hash]*types.Block),
	}
}

func sendMsgByUrl(url string, msgPayload []byte) error {
	log.Info("send msg to url %s.", url)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", url)
	if err != nil {
		log.Error("resolve tcp address %s occur fatal error: %v", url, err)
		return err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Error("dial tcp with %s occur error: %s", url, err)
		return err
	}
	log.Info("connect success, send to url %s with payload %x.", url, msgPayload)
	conn.Write(msgPayload)
	return nil
}

func (instance *bftCore) broadcast(msgPayload []byte, msgType messages.MessageType, digest types.Hash) {
	peers := instance.peers
	for id, peer := range peers {
		log.Info("broadcast from node %d to %d by url %s with message type %v and digest %x.",
			instance.local.Extension.Id, id, peer.Extension.Url, msgType, digest)
		err := sendMsgByUrl(peer.Extension.Url, msgPayload)
		if nil != err {
			log.Error("broadcast from node %d to %d by url %s with message type %v and digest %x occur error %v.",
				instance.local.Extension.Id, id, peer.Extension.Url, msgType, digest, err)
		}
	}
}

func (instance *bftCore) broadcastByOrder(msgPayload []byte, msgType messages.MessageType, digest types.Hash, peers []account.Account) {
	for _, peer := range peers {
		log.Info("broadcast from node %d to %d by url %s with message type %v and digest %x.",
			instance.local.Extension.Id, peer.Extension.Id, peer.Extension.Url, msgType, digest)
		err := sendMsgByUrl(peer.Extension.Url, msgPayload)
		if nil != err {
			log.Error("broadcast from node %d to %d by url %s with message type %v and digest %x occur error %v.",
				instance.local.Extension.Id, peer.Extension.Id, peer.Extension.Url, msgType, digest, err)
		}
	}
}

func (instance *bftCore) unicast(account account.Account, msgPayload []byte, msgType messages.MessageType, digest types.Hash) error {
	log.Info("node %d send msg [type %v, digest %x] to %d with url %s.",
		instance.local.Extension.Id, msgType, digest, account.Extension.Id, account.Extension.Url)
	err := sendMsgByUrl(account.Extension.Url, msgPayload)
	if nil != err {
		log.Error("node %d send msg [type %v and digest %x] to %d with url %s occurs error %v.",
			instance.local.Extension.Id, msgType, digest, account.Extension.Id, account.Extension.Url, err)
	}
	return err
}

func (instance *bftCore) receiveRequest(request *messages.Request) {
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
	receipts, err := instance.verifyPayload(request.Payload)
	if nil != err {
		log.Error("proposal verified failed with error %v.", err)
		return
	}
	signData, err := instance.signPayload(request.Payload.Header.MixDigest)
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
	proposal := messages.Message{
		MessageType: messages.ProposalMessageType,
		PayLoad: &messages.ProposalMessage{
			Proposal: &messages.Proposal{
				Account:   instance.local,
				Timestamp: request.Timestamp,
				Payload:   request.Payload,
				Signature: signData,
			},
		},
	}
	msgRaw, err := messages.EncodeMessage(proposal)
	if nil != err {
		log.Error("marshal proposal msg failed with %v.", err)
		return
	}
	instance.digest = request.Payload.Header.MixDigest
	instance.signature.addSignature(instance.local, signData)
	log.Info("broadcast proposal to peers.")
	instance.broadcast(msgRaw, messages.ProposalMessageType, instance.digest)
	go instance.waitResponse()
}

func (instance *bftCore) waitResponse() {
	log.Warn("set timer with 5 second.")
	timer := time.NewTimer(5 * time.Second)
	for {
		select {
		case <-timer.C:
			log.Info("response timeout.")
			signatures, err := instance.maybeCommit()
			if nil != err {
				log.Warn("maybe commit errors %s.", err)
			}
			instance.mutex.Lock()
			instance.commit = true
			instance.mutex.Unlock()
			consensusResult := &messages.ConsensusResult{
				Signatures: signatures,
				Result:     err,
			}
			instance.result <- consensusResult
			return
		case <-instance.tunnel:
			log.Info("receive tunnel")
			signatures, err := instance.maybeCommit()
			if len(signatures) == len(instance.peers) {
				instance.mutex.Lock()
				instance.commit = true
				instance.mutex.Unlock()
				consensusResult := messages.ConsensusResult{
					Signatures: signatures,
					Result:     err,
				}
				instance.result <- &consensusResult
				log.Info("receive all response before timeout")
				return
			}
			log.Warn("get %d signatures of %d peers.", len(signatures), len(instance.peers))
		}
	}
}

func (instance *bftCore) receiveProposal(proposal *messages.Proposal) {
	isMaster := instance.local == instance.master
	if isMaster {
		log.Info("master not need to process proposal.")
		return
	}
	if instance.master != proposal.Account {
		log.Error("proposal must from master %d, while it from %d in fact.", instance.master, proposal.Account.Extension.Id)
		return
	}
	if !signDataVerify(instance.master, proposal.Signature, proposal.Payload.Header.MixDigest) {
		log.Error("proposal signature not from master, please confirm.")
		return
	}
	receipts, err := instance.verifyPayload(proposal.Payload)
	if nil != err {
		log.Error("proposal verified failed with error %v.", err)
		return
	}
	signData, err := instance.signPayload(proposal.Payload.Header.MixDigest)
	if nil != err {
		log.Error("archive proposal signature failed with error %v.", err)
		return
	}
	// ensure reserve receipts must be verified and signed
	if values, ok := instance.validator[proposal.Payload.Header.MixDigest]; !ok {
		log.Info("add record payload %x.", proposal.Payload.Header.MixDigest)
		instance.validator[proposal.Payload.Header.MixDigest] = &payloadSets{
			block:    proposal.Payload,
			receipts: receipts,
		}
	} else {
		values.receipts = receipts
	}
	response := messages.Message{
		MessageType: messages.ResponseMessageType,
		PayLoad: &messages.ResponseMessage{
			Response: &messages.Response{
				Account:   instance.local,
				Timestamp: proposal.Timestamp,
				Digest:    proposal.Payload.Header.MixDigest,
				Signature: signData,
			},
		},
	}
	msgRaw, err := messages.EncodeMessage(response)
	if nil != err {
		log.Error("marshal proposal msg failed with %v.", err)
		return
	}
	err = instance.unicast(instance.master, msgRaw, messages.ResponseMessageType, proposal.Payload.Header.MixDigest)
	if err != nil {
		log.Error("unicast to master %x failed with error %v.", instance.master.Address, err)
	}
}

func (instance *bftCore) verifyPayload(payload *types.Block) (types.Receipts, error) {
	blockStore, err := blockchain.NewBlockChainByBlockHash(payload.Header.PrevBlockHash)
	if nil != err {
		log.Error("Get NewBlockChainByBlockHash failed.")
		return nil, err
	}
	worker := worker.NewWorker(blockStore, payload, true)
	err = worker.VerifyBlock()
	if err != nil {
		log.Error("The block %d verified failed with err %v.", payload.Header.Height, err)
		return nil, err
	}

	return worker.GetReceipts(), nil
}

func (instance *bftCore) signPayload(digest types.Hash) ([]byte, error) {
	sign, err := signature.Sign(&instance.local, digest[:])
	if nil != err {
		log.Error("archive signature occur error %x.", err)
		return nil, err
	}
	log.Info("archive signature for %x successfully with sign %x.", digest, sign)
	return sign, nil
}

func (instance *bftCore) maybeCommit() ([][]byte, error) {
	var reallySignature = make([][]byte, 0)
	if uint8(len(instance.signature.signatures)) < uint8(len(instance.peers))-instance.tolerance {
		log.Info("commit need %d signature, while now is %d.",
			uint8(len(instance.peers))-instance.tolerance, len(instance.signature.signatures))
	}
	signData := instance.signature.signatures
	signMap := instance.signature.signMap
	if len(signData) != len(signMap) {
		log.Error("length of signData[%d] and signMap[%d] does not match.", len(signData), len(signMap))
		return reallySignature, fmt.Errorf("signData and signMap does not match")
	}
	var suspiciousAccount = make([]account.Account, 0)
	for account, sign := range signMap {
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
		log.Error("verify sign %v failed with err %s", sign, err)
	}
	return account.Address == address
}

func (instance *bftCore) receiveResponse(response *messages.Response) {
	instance.mutex.RLock()
	defer instance.mutex.RUnlock()
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
		peer := tools.GetAccountById(instance.peers, response.Account.Extension.Id)
		if !signDataVerify(peer, response.Signature, instance.digest) {
			log.Error("signature and response sender not in coincidence.")
			return
		}
		if sign, ok := instance.signature.signMap[peer]; !ok {
			instance.signature.addSignature(peer, response.Signature)
			log.Info("try to notify toCommit.")
			instance.tunnel <- 1
			log.Info("response from %x as been committed success.", response.Account.Address)
		} else {
			// check signature
			if !bytes.Equal(sign, response.Signature) {
				log.Error("receive a different signature from the same validator %x, which exists is %x, while response is %x.",
					peer.Address, sign, response.Signature)
			}
			log.Warn("receive duplicate signature from the same validator, ignore it.")
		}
		return
	} else {
		log.Info("response has be committed, ignore response from %x.", response.Account.Address)
		return
	}
}

func (instance *bftCore) getCommitOrder(result error, currentMaster int) []account.Account {
	var nextMaster = -1
	peers := make([]account.Account, 0)
	if nil == result {
		peers = append(peers, instance.peers[currentMaster])
		nextMaster = (1 + currentMaster) % len(instance.peers)
	}
	for index, accounts := range instance.peers {
		if index != currentMaster && index != nextMaster {
			peers = append(peers, accounts)
		}
	}
	if -1 != nextMaster {
		peers = append(peers, instance.peers[nextMaster])
	} else {
		peers = append(peers, instance.peers[currentMaster])
	}
	log.Info("commit order %v", peers)
	return peers
}

func (instance *bftCore) commitFilter(blacklist account.Account) []account.Account {
	peers := make([]account.Account, 0)
	for index, accounts := range instance.peers {
		if index != int(blacklist.Extension.Id) {
			peers = append(peers, accounts)
		}
	}
	log.Info("commit order %v", peers)
	return peers
}

func (instance *bftCore) SendCommit(commit *messages.Commit, block *types.Block) {
	committed := messages.Message{
		MessageType: messages.CommitMessageType,
		PayLoad: &messages.CommitMessage{
			Commit: commit,
		},
	}
	msgRaw, err := messages.EncodeMessage(committed)
	if nil != err {
		log.Error("marshal commit msg failed with %v.", err)
		return
	}
	if !commit.Result {
		peers := instance.commitFilter(instance.local)
		instance.broadcastByOrder(msgRaw, messages.CommitMessageType, commit.Digest, peers)
		log.Info("later to notify local")
		instance.eventCenter.Notify(types.EventConsensusFailed, nil)
	} else {
		log.Info("first to notify local")
		nextMaster := int(instance.local.Extension.Id+1) % len(instance.peers)
		peers := make([]account.Account, 0)
		for index, accounts := range instance.peers {
			if index != nextMaster && index != int(instance.local.Extension.Id) {
				peers = append(peers, accounts)
			}
		}
		peers = append(peers, instance.peers[nextMaster])
		instance.commitBlock(block)
		instance.broadcastByOrder(msgRaw, messages.CommitMessageType, commit.Digest, peers)
	}
}

func (instance *bftCore) receiveCommit(commit *messages.Commit) {
	log.Info("receive commit")
	if !commit.Result {
		log.Error("receive commit with error %v.", commit.Result)
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

func (instance *bftCore) commitBlock(block *types.Block) {
	chain, err := blockchain.NewBlockChainByBlockHash(block.Header.PrevBlockHash)
	if nil != err {
		block.Header.SigData = make([][]byte, 0)
		log.Error("get NewBlockChainByHash by hash %x failed with error %s.", block.Header.PrevBlockHash, err)
		return
	}
	block.HeaderHash = common.HeaderHash(block)
	log.Info("begin write block %d with hash %x.", block.Header.Height, block.HeaderHash)
	err = chain.WriteBlockWithReceipts(block, instance.validator[block.Header.MixDigest].receipts)
	if nil != err {
		block.Header.SigData = make([][]byte, 0)
		log.Error("call WriteBlockWithReceipts failed with", block.Header.PrevBlockHash, err)
	}
	log.Info("end write block %d with hash %x with success.", block.Header.Height, block.HeaderHash)
}

func (instance *bftCore) ProcessEvent(e tools.Event) tools.Event {
	var err error
	log.Debug("replica %d processing event", instance.local.Extension.Id)
	switch et := e.(type) {
	case *messages.Request:
		log.Info("receive request from replica %d.", instance.local.Extension.Id)
		instance.receiveRequest(et)
	case *messages.Proposal:
		log.Info("receive proposal from replica %d with digest %x.", et.Account.Extension.Id, et.Payload.Header.MixDigest)
		instance.receiveProposal(et)
	case *messages.Response:
		log.Info("receive response from replica %d with digest %x.", et.Account.Extension.Id, et.Digest)
		instance.receiveResponse(et)
	case *messages.Commit:
		log.Info("receive commit from replica %d with digest %x.", et.Account.Extension.Id, et.Digest)
		instance.receiveCommit(et)
	default:
		log.Warn("replica %d received an unknown message type %T", instance.local.Extension.Id, et)
		err = fmt.Errorf("un support type %v", et)
	}
	if err != nil {
		log.Warn(err.Error())
	}
	return err
}

func (instance *bftCore) Start(account account.Account) {
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
	handleConnection(tcpListener, instance)
}

func handleConnection(tcpListener *net.TCPListener, bft *bftCore) {
	for {
		var conn, _ = tcpListener.AcceptTCP()
		reader := bufio.NewReaderSize(conn, common.MaxBufferLen)
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
				proposal.Account.Extension.Id, proposal.Payload.Header.MixDigest)
			if proposal.Account != bft.master {
				log.Warn("only master can issue a proposal.")
				continue
			}
			tools.SendEvent(bft, proposal)
		case messages.ResponseMessageType:
			response := payload.(*messages.ResponseMessage).Response
			log.Info("receive response message from node %d with payload %x.",
				response.Account.Extension.Id, response.Digest)
			if response.Account.Extension.Id == bft.master.Extension.Id {
				log.Warn("master will not receive response message from itself.")
				continue
			}
			tools.SendEvent(bft, response)
		case messages.CommitMessageType:
			commit := payload.(*messages.CommitMessage).Commit
			tools.SendEvent(bft, commit)
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
