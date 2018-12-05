package messages

import (
	"encoding/json"
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/validator/tools/account"
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

type SyncBlockReq struct {
	Node       account.Account
	Timestamp  int64
	BlockStart uint64
	BlockEnd   uint64
}

type SyncBlockResp struct {
	// TODO: add signatures
	Blocks []*types.Block
}

type ViewChangeReq struct {
	Id        uint64
	Nodes     []account.Account
	Timestamp int64
	ViewNum   uint64
}

type MessageType string

var RequestMessageType MessageType = "RequestMessage"
var ProposalMessageType MessageType = "ProposalMessage"
var ResponseMessageType MessageType = "ResponseMessage"
var CommitMessageType MessageType = "CommitMessage"
var SyncBlockReqMessageType MessageType = "SyncBlockReqMessage"
var SyncBlockRespMessageType MessageType = "SyncBlockResMessage"
var ViewChangeMessageReqType MessageType = "ViewChangeReqMessage"

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

type SyncBlockReqMessage struct {
	SyncBlock *SyncBlockReq
}

type SyncBlockRespMessage struct {
	SyncBlock *SyncBlockResp
}

type ViewChangeReqMessage struct {
	ViewChange *ViewChangeReq
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
	err = json.Unmarshal(rawData, &messageClone)
	if err != nil {
		log.Error("default unmarshal messageClone failed with error %v.", err)
		return err
	}
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
	case SyncBlockReqMessageType:
		payload := &SyncBlockReqMessage{}
		err = json.Unmarshal(messageClone.Payload, payload)
		if nil != err {
			log.Error("unmarshal sync block request message failed with err %v.", err)
		}
		m.Payload = payload
	case SyncBlockRespMessageType:
		payload := &SyncBlockRespMessage{}
		err = json.Unmarshal(messageClone.Payload, payload)
		if nil != err {
			log.Error("unmarshal sync block response message failed with err %v.", err)
		}
		m.Payload = payload
	case ViewChangeMessageReqType:
		payload := &ViewChangeReqMessage{}
		err = json.Unmarshal(messageClone.Payload, payload)
		if nil != err {
			log.Error("unmarshal view change request message failed with err %v.", err)
		}
		m.Payload = payload
	default:
		log.Error("not support marshal type %v.", m.MessageType)
		err = fmt.Errorf("not support marshal type")
	}
	return err
}
