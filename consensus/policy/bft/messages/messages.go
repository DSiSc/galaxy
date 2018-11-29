package messages

import (
	"encoding/json"
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/validator/tools/account"
	"net"
)

type ConsensusResult struct {
	Signatures [][]byte
	Result     error
}

// messages of events
type Request struct {
	Timestamp int64
	Payload   *types.Block
}

type Proposal struct {
	Id        uint64
	Timestamp int64
	Payload   *types.Block
	Signature []byte
}

type Response struct {
	Account   account.Account
	Timestamp int64
	Digest    types.Hash
	Signature []byte
}

type Commit struct {
	Account    account.Account
	Timestamp  int64
	BlockHash  types.Hash
	Digest     types.Hash
	Signatures [][]byte
	Result     bool
}

type SyncBlock struct {
	Node account.Account
	Timestamp  int64
	BlockStart uint64
	BlockEnd   uint64
}

type MessageType string

var RequestMessageType MessageType = "RequestMessage"
var ProposalMessageType MessageType = "ProposalMessage"
var ResponseMessageType MessageType = "ResponseMessage"
var CommitMessageType MessageType = "CommitMessage"
var SyncBlockMessageType MessageType = "SyncBlockMessage"

type RequestMessage struct {
	Request *Request
}

type ProposalMessage struct {
	Proposal *Proposal
}

type ResponseMessage struct {
	Response *Response
}

type CommitMessage struct {
	Commit *Commit
}

type SyncBlockMessage struct {
	SyncBlock *SyncBlock
} 

type Message struct {
	MessageType MessageType
	Payload     interface{}
}

type MessageClone struct {
	MessageType MessageType
	Payload     []byte
}

func (m *Message) MarshalJSON() ([]byte, error) {
	var err error
	messageClone := MessageClone{}
	messageClone.MessageType = m.MessageType
	messageClone.Payload, err = json.Marshal(m.Payload)
	if nil != err {
		log.Error("marshal json failed with %v", err)
		return nil, fmt.Errorf("marshal json failed with")
	}

	return json.Marshal(messageClone)
}

func (m *Message) UnmarshalJSON(rawData []byte) error {
	var err error
	messageClone := MessageClone{}
	json.Unmarshal(rawData, &messageClone)
	m.MessageType = messageClone.MessageType
	switch m.MessageType {
	case RequestMessageType:
		payload := &RequestMessage{}
		err = json.Unmarshal(messageClone.Payload, payload)
		if nil != err {
			log.Error("unmarshal request message failed with err %v.", err)
		}
		m.Payload = payload
	case ProposalMessageType:
		payload := &ProposalMessage{}
		err = json.Unmarshal(messageClone.Payload, payload)
		if nil != err {
			log.Error("unmarshal proposal message failed with err %v.", err)
		}
		m.Payload = payload
	case ResponseMessageType:
		payload := &ResponseMessage{}
		err = json.Unmarshal(messageClone.Payload, payload)
		if nil != err {
			log.Error("unmarshal response message failed with err %v.", err)
		}
		m.Payload = payload
	case CommitMessageType:
		payload := &CommitMessage{}
		err = json.Unmarshal(messageClone.Payload, payload)
		if nil != err {
			log.Error("unmarshal commit message failed with err %v.", err)
		}
		m.Payload = payload
	default:
		log.Error("not support marshal type %v.", m.MessageType)
		err = fmt.Errorf("not support marshal type")
	}
	return err
}

func BroadcastPeers(msgPayload []byte, msgType MessageType, digest types.Hash, peers []account.Account) {
	for _, peer := range peers {
		log.Info("broadcast to %d by url %s with message type %v and digest %x.",
			peer.Extension.Id, peer.Extension.Url, msgType, digest)
		err := sendMsgByUrl(peer.Extension.Url, msgPayload)
		if nil != err {
			log.Error("broadcast to %d by url %s with message type %v and digest %x occur error %v.",
				peer.Extension.Id, peer.Extension.Url, msgType, digest, err)
		}
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
	_, err = conn.Write(msgPayload)
	return err
}

func Unicast(account account.Account, msgPayload []byte, msgType MessageType, digest types.Hash) error {
	log.Info("send msg [type %v, digest %x] to %d with url %s.", msgType, digest, account.Extension.Id, account.Extension.Url)
	err := sendMsgByUrl(account.Extension.Url, msgPayload)
	if nil != err {
		log.Error("send msg [type %v and digest %x] to %d with url %s occurs error %v.",
			msgType, digest, account.Extension.Id, account.Extension.Url, err)
	}
	return err
}
