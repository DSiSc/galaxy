package bft

import (
	"bytes"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/galaxy/consensus/policy/bft/messages"
	"github.com/DSiSc/galaxy/consensus/policy/bft/tools"
	"github.com/DSiSc/validator/tools/account"
	"github.com/golang/protobuf/proto"
	"net"
)

type bftCore struct {
	id        uint64
	master    uint64
	peers     []account.Account
	signature *signData
	tolerance uint8
	result    chan *signData
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

func NewBFTCore(id uint64) *bftCore {
	return &bftCore{
		id: id,
		signature: &signData{
			signatures: make([][]byte, 0),
			signMap:    make(map[account.Account][]byte),
		},
		result: make(chan *signData),
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
	// send
	isMaster := instance.id == instance.master
	if !isMaster {
		log.Info("only master process request.")
		return
	}
	signature := request.Payload.Header.SigData
	if 1 != len(signature) {
		log.Error("request must have signature from client.")
		return
	}
	// TODO: Add master signature to proposal
	// now, client is master, so we just used client's signature
	master := instance.peers[instance.master]
	instance.signature.addSignature(master, signature[0])
	proposal := &messages.Message{
		Payload: &messages.Message_Proposal{
			Proposal: &messages.Proposal{
				Id:        instance.id,
				Timestamp: request.Timestamp,
				Payload:   request.Payload,
				Signature: signature[0],
			},
		},
	}
	msgRaw, err := proto.Marshal(proposal)
	if nil != err {
		log.Error("marshal proposal msg failed with %v.", err)
		return
	}
	instance.broadcast(msgRaw)
}

func (instance *bftCore) receiveProposal(proposal *messages.Proposal) {
	isMaster := instance.id == instance.master
	if isMaster {
		log.Info("master not need to process proposal.")
		return
	}
	// TODO: Add signature
	response := &messages.Message{
		Payload: &messages.Message_Response{
			Response: &messages.Response{
				Id:        instance.id,
				Timestamp: proposal.Timestamp,
				// TODO: add sign to the block
				Payload: proposal.Payload,
			},
		},
	}
	msgRaw, err := proto.Marshal(response)
	if nil != err {
		log.Error("marshal proposal msg failed with %v.", err)
		return
	}
	masterAccount := instance.peers[instance.master]
	err = instance.unicast(masterAccount, msgRaw)
	if err != nil {
		log.Error("unicast to master %x failed with error %v.", masterAccount.Address, err)
	}
}

func (instance *bftCore) maybeCommit() {
	signatures := len(instance.signature.signatures)
	if uint8(signatures) < instance.tolerance {
		log.Info("commit need %d signature, while now is %d.", instance.tolerance, signatures)
		return
	}
	log.Info("commit it.")
	// TODO: send message to ToConsensus
	// instance.result <- instance.signature
}

func (instance *bftCore) receiveResponse(response *messages.Response) {
	isMaster := instance.id == instance.master
	if !isMaster {
		log.Info("only master need to process response.")
		return
	}
	peer := instance.peers[response.Id]
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
	instance.maybeCommit()
}

func (instance *bftCore) ProcessEvent(e tools.Event) tools.Event {
	var err error
	log.Debug("replica %d processing event", instance.id)
	switch et := e.(type) {
	case *messages.Request:
		log.Info("receive request from replica %d.", instance.id)
		instance.receiveRequest(et)
	case *messages.Proposal:
		log.Info("receive proposal from replica %d.", et.Id)
		instance.receiveProposal(et)
	case *messages.Response:
		log.Info("receive proposal from replica %d.", et.Id)
		instance.receiveResponse(et)
	default:
		log.Warn("replica %d received an unknown message type %T", instance.id, et)
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
		err = proto.Unmarshal(buffer[:n], &msg)
		payload := msg.GetPayload()
		switch payload.(type) {
		case *messages.Message_Request:
			log.Info("receive request message.")
			request := payload.(*messages.Message_Request).Request
			tools.SendEvent(bft, request)
		case *messages.Message_Proposal:
			log.Info("receive proposal message.")
			proposal := payload.(*messages.Message_Proposal).Proposal
			tools.SendEvent(bft, proposal)
		case *messages.Message_Response:
			log.Info("receive response message.")
			response := payload.(*messages.Message_Response).Response
			tools.SendEvent(bft, response)
		default:
			log.Error("not support type for handleConnection.")
			continue
		}
	}
}
