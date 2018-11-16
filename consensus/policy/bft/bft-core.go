package bft

import (
	"bytes"
	"encoding/json"
	"github.com/DSiSc/blockchain"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus/policy/bft/messages"
	"github.com/DSiSc/galaxy/consensus/policy/bft/tools"
	"github.com/DSiSc/validator/tools/account"
	"github.com/DSiSc/validator/tools/signature"
	"github.com/DSiSc/validator/worker"
	"net"
)

type bftCore struct {
	local     account.Account
	master    uint64
	peers     []account.Account
	signature *signData
	tolerance uint8
	commit    bool
	digest    types.Hash
	result    chan messages.SignatureSet
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

func NewBFTCore(local account.Account, result chan messages.SignatureSet) *bftCore {
	return &bftCore{
		local: local,
		signature: &signData{
			signatures: make([][]byte, 0),
			signMap:    make(map[account.Account][]byte),
		},
		result: result,
	}
}

func sendMsgByUrl(url string, msgPayload []byte) error {
	log.Info("send msg to  url %s.\n", url)
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

func (instance *bftCore) broadcast(msgPayload []byte) {
	peers := instance.peers
	for id, peer := range peers {
		log.Info("broadcast to node %d with url %s.", id, peer.Extension.Url)
		err := sendMsgByUrl(peer.Extension.Url, msgPayload)
		if nil != err {
			log.Error("broadcast to %s error %s.", peer.Extension.Url, err)
		}
	}
}

func (instance *bftCore) unicast(account account.Account, msgPayload []byte) error {
	log.Info("send msg to node %d with url %s.\n", account.Extension.Id, account.Extension.Url)
	err := sendMsgByUrl(account.Extension.Url, msgPayload)
	if nil != err {
		log.Error("send msg to url %s failed with %v.", account.Extension.Url, err)
	}
	return err
}

func (instance *bftCore) receiveRequest(request *messages.Request) {
	isMaster := instance.local.Extension.Id == instance.master
	if !isMaster {
		log.Info("only master process request.")
		return
	}
	signature := request.Payload.Header.SigData
	if 1 != len(signature) {
		log.Error("request must have signature from producer.")
		return
	}
	err := instance.verifyPayload(request.Payload)
	if nil != err {
		log.Error("proposal verified failed with error %v.", err)
		return
	}
	signData, err := instance.signPayload(request.Payload.Header.MixDigest)
	if nil != err {
		log.Error("archive proposal signature failed with error %v.", err)
		return
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
	instance.signature.addSignature(instance.local, signData)
	instance.broadcast(msgRaw)
}

func (instance *bftCore) receiveProposal(proposal *messages.Proposal) {
	isMaster := instance.local.Extension.Id == instance.master
	if isMaster {
		log.Info("master not need to process proposal.")
		return
	}
	if instance.master != proposal.Id {
		log.Error("proposal must from master %d, while it from %d in fact.", instance.master, proposal.Id)
		return
	}
	masterAccount := instance.peers[instance.master]
	if !signDataVerify(masterAccount, proposal.Signature, proposal.Payload.Header.MixDigest) {
		log.Error("proposal signature not from master, please confirm.")
		return
	}
	err := instance.verifyPayload(proposal.Payload)
	if nil != err {
		log.Error("proposal verified failed with error %v.", err)
		return
	}
	signData, err := instance.signPayload(proposal.Payload.Header.MixDigest)
	if nil != err {
		log.Error("archive proposal signature failed with error %v.", err)
		return
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
	err = instance.unicast(masterAccount, msgRaw)
	if err != nil {
		log.Error("unicast to master %x failed with error %v.", masterAccount.Address, err)
	}
}

func (instance *bftCore) verifyPayload(payload *types.Block) error {
	blockStore, err := blockchain.NewBlockChainByBlockHash(payload.Header.PrevBlockHash)
	if nil != err {
		log.Error("Get NewBlockChainByBlockHash failed.")
		return err
	}
	worker := worker.NewWorker(blockStore, payload)
	err = worker.VerifyBlock()
	if err != nil {
		log.Error("The block %d verified failed with err %v.", payload.Header.Height, err)
		return err
	}
	return nil
}

func (instance *bftCore) signPayload(digest types.Hash) ([]byte, error) {
	sign, err := signature.Sign(&instance.local, digest[:])
	if nil != err {
		log.Error("archive signature error.")
		return nil, err
	}
	log.Info("archive signature for %x successfully.", digest)
	return sign, nil
}

func (instance *bftCore) maybeCommit() {
	if uint8(len(instance.signature.signatures)) < uint8(len(instance.peers))-instance.tolerance {
		log.Info("commit need %d signature, while now is %d.",
			uint8(len(instance.peers))-instance.tolerance, len(instance.signature.signatures))
	}
	signData := instance.signature.signatures
	signMap := instance.signature.signMap
	if len(signData) != len(signMap) {
		log.Error("length of signData[%d] and signMap[%d] does not match.", len(signData), len(signMap))
		return
	}
	var reallySignature = make([][]byte, 0)
	var suspiciousAccount = make([]account.Account, 0)
	for account, sign := range signMap {
		if signDataVerify(account, sign, instance.digest) {
			reallySignature = append(reallySignature, sign)
			continue
		}
		suspiciousAccount = append(suspiciousAccount, account)
		log.Error("signature %x by account %x is invalid", sign, account)
	}
	if uint8(len(reallySignature)) < uint8(len(instance.peers))-instance.tolerance {
		log.Error("really signature %d less than need %d.",
			len(reallySignature), uint8(len(instance.peers))-instance.tolerance)
		return
	}
	instance.commit = true
	instance.result <- reallySignature
}

func signDataVerify(account account.Account, sign []byte, digest types.Hash) bool {
	address, err := signature.Verify(digest, sign)
	if nil != err {
		log.Error("verify sign %v failed with err %s", sign, err)
	}
	return account.Address == address
}

func (instance *bftCore) receiveResponse(response *messages.Response) {
	isMaster := instance.local.Extension.Id == instance.master
	if !isMaster {
		log.Info("only master need to process response.")
		return
	}
	if !bytes.Equal(instance.digest[:], response.Digest[:]) {
		log.Error("received response digest %x not in coincidence with reserved %x.",
			instance.digest, response.Digest)
		return
	}
	peer := instance.peers[response.Account.Extension.Id]
	if !signDataVerify(peer, response.Signature, instance.digest) {
		log.Error("signature and response sender not in coincidence.")
		return
	}
	if sign, ok := instance.signature.signMap[peer]; !ok {
		instance.signature.addSignature(peer, response.Signature)
	} else {
		// check signature
		if !bytes.Equal(sign, response.Signature) {
			log.Error("receive a different signature from the same validator %x, which exists is %x, while response is %x.",
				peer.Address, sign, response.Signature)
			return
		}
		log.Warn("receive duplicate signature from the same validator, ignore it.")
	}
	if !instance.commit {
		log.Info("try to commit the response.")
		instance.maybeCommit()
	} else {
		log.Warn("response has been committed already.")
	}
}

func (instance *bftCore) ProcessEvent(e tools.Event) tools.Event {
	var err error
	log.Debug("replica %d processing event", instance.local.Extension.Id)
	switch et := e.(type) {
	case *messages.Request:
		log.Info("receive request from replica %d.", instance.local.Extension.Id)
		instance.receiveRequest(et)
	case *messages.Proposal:
		log.Info("receive proposal from replica %d.", et.Id)
		instance.receiveProposal(et)
	case *messages.Response:
		log.Info("receive response from replica %d.", et.Account.Extension.Id)
		instance.receiveResponse(et)
	default:
		log.Warn("replica %d received an unknown message type %T", instance.local.Extension.Id, et)
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
	buffer := make([]byte, 2048)
	for {
		var conn, _ = tcpListener.AcceptTCP()
		n, err := conn.Read(buffer)
		if err != nil {
			log.Error("error when read connector %x.", err)
			return
		}
		log.Info("receive messages form other node.")
		var msg messages.Message
		err = json.Unmarshal(buffer[:n], &msg)
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
			if proposal.Id != bft.master {
				log.Warn("only master can issue a proposal.", bft.master)
				continue
			}
			tools.SendEvent(bft, proposal)
		case messages.ResponseMessageType:
			response := payload.(*messages.ResponseMessage).Response
			log.Info("receive response message from node %d with payload %x.",
				response.Account.Extension.Id, response.Digest)
			if response.Account.Extension.Id == bft.master {
				log.Warn("master %d will not receive response message from itself.")
				continue
			}
			tools.SendEvent(bft, response)
		default:
			if nil == payload {
				log.Info("receive handshake, omit it.")
			} else {
				log.Error("not support type for %v.", payload)
			}
			continue
		}
	}
}
