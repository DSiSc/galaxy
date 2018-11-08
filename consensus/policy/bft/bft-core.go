package bft

import (
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/galaxy/consensus/policy/bft/messages"
	"github.com/DSiSc/galaxy/consensus/policy/bft/tools"
	"github.com/DSiSc/validator/tools/account"
	"github.com/golang/protobuf/proto"
	"net"
)

type bftCore struct {
	id      uint64
	manager *server
}

type server struct {
	url    string
	listen *net.TCPListener
}

func NewBFTCore(id uint64) tools.Receiver {
	return &bftCore{
		id: id,
	}
}

func (instance *bftCore) ProcessEvent(e tools.Event) tools.Event {
	var err error
	log.Debug("replica %d processing event", instance.id)
	switch et := e.(type) {
	case messages.Request:
		log.Info("receive request from replica %d.", instance.id)
	case messages.Proposal:
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
	instance.manager = &server{
		listen: tcpListener,
		url:    url,
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
