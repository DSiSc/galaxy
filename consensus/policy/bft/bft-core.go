package bft

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/galaxy/consensus/policy/bft/messages"
	"github.com/DSiSc/galaxy/consensus/policy/bft/tools"
	"github.com/DSiSc/validator/tools/account"
	"github.com/golang/protobuf/proto"
	"net"
	"os"
)

type bftCore struct {
	id       uint64
	isMaster bool
	peers    []account.Account
}

func NewBFTCore(id uint64, master bool, peers []account.Account) tools.Receiver {
	return &bftCore{
		id:       id,
		isMaster: master,
		peers:    peers,
	}
}

func (instance *bftCore) broadcast(msgPayload []byte) {
	peers := instance.peers
	for id, peer := range peers {
		log.Info("Broadcast to node %d with url %s.\n", id, peer.Extension.Url)
		tcpAddr, err := net.ResolveTCPAddr("tcp4", peer.Extension.Url)
		if err != nil {
			log.Error("Fatal error: %s", err.Error())
			continue
		}
		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
			continue
		}
		log.Info("connect success and  send to url %s and payload %x.\n", peer.Extension.Url, msgPayload)
		conn.Write(msgPayload)
	}
}

func (instance *bftCore) receiveRequest(request *messages.Request) {
	// send
	proposal := &messages.Proposal{
		Id:        instance.id,
		Timestamp: request.Timestamp,
		Payload:   request.Payload,
	}
	porposalMsg := &messages.Message{
		Payload: &messages.Message_Proposal{
			Proposal: proposal,
		},
	}
	msgRaw, err := proto.Marshal(porposalMsg)
	if nil != err {
		log.Error("marshal proposal msg failed with %v.", err)
		return
	}
	instance.broadcast(msgRaw)
}

func (instance *bftCore) ProcessEvent(e tools.Event) tools.Event {
	var err error
	log.Debug("replica %d processing event", instance.id)
	switch et := e.(type) {
	case *messages.Request:
		log.Info("receive request from replica %d.", instance.id)
		instance.receiveRequest(et)
	case *messages.Proposal:
		log.Info("receive proposal from replica %d.", instance.id)
	default:
		log.Warn("replica %d received an unknown message type %T", instance.id, et)
	}
	if err != nil {
		log.Warn(err.Error())
	}

	return nil
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
			request := payload.(*messages.Message_Request).Request
			tools.SendEvent(bft, request)
		case *messages.Message_Proposal:
			proposal := payload.(*messages.Message_Proposal).Proposal
			tools.SendEvent(bft, proposal)
		default:
			log.Error("not support type for handleConnection.")
			continue
		}
	}
}
