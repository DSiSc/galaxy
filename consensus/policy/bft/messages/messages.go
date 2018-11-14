package messages

import (
	"encoding/json"
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
)

type SignatureSet [][]byte

// messages of events
type Request struct {
	Timestamp int64
	Payload   *types.Block
}

type Proposal struct {
	Timestamp int64
	Payload   *types.Block
	Id        uint64
	Signature []byte
}

type Response struct {
	Id        uint64
	Timestamp int64
	// TODO: used sig to replace Payload
	Payload   *types.Block
	Signature []byte
}

type MessageType string

var RequestMessageType MessageType = "RequestMessage"
var ProposalMessageType MessageType = "ProposalMessage"
var ResponseMessageType MessageType = "ResponseMessage"

type RequestMessage struct {
	Request *Request
}

type ProposalMessage struct {
	Proposal *Proposal
}

type ResponseMessage struct {
	Response *Response
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
	default:
		log.Error("uot support marshal type %v.", m.MessageType)
		err = fmt.Errorf("not support marshal type")
	}
	return err
}
