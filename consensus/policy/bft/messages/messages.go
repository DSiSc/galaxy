package messages

import (
	"github.com/DSiSc/craft/types"
	"github.com/golang/protobuf/proto"
)

// messages of events
type Request struct {
	Timestamp int64
	Payload   *types.Block
}

type Proposal struct {
	Timestamp int64
	Payload   *types.Block
	Id        uint64
}

// message of nodes
type Message struct {
	Payload isMessage_Payload `protobuf_oneof:"payload"`
}

func (m *Message) Reset()                    { *m = Message{} }
func (m *Message) String() string            { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()               {}
func (*Message) Descriptor() ([]byte, []int) { return nil, []int{0} }

type isMessage_Payload interface {
	isMessage_Payload()
}

func (m *Message) GetPayload() isMessage_Payload {
	if m != nil {
		return m.Payload
	}
	return nil
}

type Message_Request struct {
	Request *Request
}

func (*Message_Request) isMessage_Payload() {}

type Message_Proposal struct {
	Proposal *Proposal
}

func (*Message_Proposal) isMessage_Payload() {}
