package messages

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/validator/tools/account"
	"io"
	"net"
)

//MessageType is the message type
type MessageType uint32

const (
	_                        = MessageType(iota) // 0, nil ,message
	RequestMessageType                           // 1, request message for bft
	ProposalMessageType                          // 2, proposal message for bft
	ResponseMessageType                          // 3, response message for bft
	CommitMessageType                            // 4, commit message for bft
	SyncBlockReqMessageType                      // 5, sync block request
	SyncBlockRespMessageType                     // 6, sync block response
	ViewChangeMessageReqType                     // 7, change view request
	OnlineRequestType                            // 8, node online request
	OnlineResponseType                           // 9, response for node online request
)

type Message struct {
	MessageType MessageType
	PayLoad     interface{}
}

type MessageHeader struct {
	Magic       uint32
	MessageType MessageType
	Length      uint32
}

// EncodeMessage encode message to byte array.
func EncodeMessage(msg Message) ([]byte, error) {
	msgByte, err := json.Marshal(msg.PayLoad)
	if err != nil {
		return nil, fmt.Errorf("failed to encode message %v to json, as: %v", msg, err)
	}

	header, err := buildMessageHeader(msg, len(msgByte))
	if err != nil {
		return nil, err
	}

	buf, err := encodeMessageHeader(header)
	if err != nil {
		return nil, err
	}

	return append(buf, msgByte...), nil
}

// fill the header according to the message.
func buildMessageHeader(msg Message, len int) (*MessageHeader, error) {
	header := &MessageHeader{
		Magic:       0,
		MessageType: msg.MessageType,
		Length:      uint32(len),
	}
	return header, nil
}

// encodeMessageHeader encode message header to byte array.
func encodeMessageHeader(header *MessageHeader) ([]byte, error) {
	buf := make([]byte, 12)
	binary.LittleEndian.PutUint32(buf, header.Magic)
	binary.LittleEndian.PutUint32(buf[4:], uint32(header.MessageType))
	binary.LittleEndian.PutUint32(buf[8:], header.Length)
	return buf, nil
}

// ReadMessage read message
func ReadMessage(reader io.Reader) (Message, error) {
	header, err := readMessageHeader(reader)
	if err != nil {
		return Message{}, err
	}

	body := make([]byte, header.Length)
	_, err = io.ReadFull(reader, body)
	if err != nil {
		return Message{}, err
	}

	return DecodeMessage(header.MessageType, body)
}

// read message header from reader.
func readMessageHeader(reader io.Reader) (MessageHeader, error) {
	header := MessageHeader{}
	err := binary.Read(reader, binary.LittleEndian, &header)
	return header, err
}

// make empty message according to the message type
func makeEmptyMessage(MessageType MessageType) (interface{}, error) {
	switch MessageType {
	case RequestMessageType:
		return &RequestMessage{}, nil
	case ProposalMessageType:
		return &ProposalMessage{}, nil
	case ResponseMessageType:
		return &ResponseMessage{}, nil
	case CommitMessageType:
		return &CommitMessage{}, nil
	case SyncBlockReqMessageType:
		return &SyncBlockReqMessage{}, nil
	case SyncBlockRespMessageType:
		return &SyncBlockRespMessage{}, nil
	case ViewChangeMessageReqType:
		return &ViewChangeReqMessage{}, nil
	case OnlineRequestType:
		return &OnlineRequestMessage{}, nil
	case OnlineResponseType:
		return &OnlineResponseMessage{}, nil
	default:
		return nil, fmt.Errorf("unknown message type %v", MessageType)
	}
}

func DecodeMessage(MessageType MessageType, rawMsg []byte) (Message, error) {
	payload, err := makeEmptyMessage(MessageType)
	if nil != err {
		return Message{}, err
	}
	err = json.Unmarshal(rawMsg, payload)
	if nil != err {
		log.Error("unmarshal rawMsg failed with err %v.", err)
		return Message{}, err
	}
	return Message{MessageType: MessageType, PayLoad: payload}, nil
}

func sendMsgByUrl(url string, msgPayload []byte) error {
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
	defer conn.Close()
	_, err = conn.Write(msgPayload)
	if nil != err {
		log.Error("write connection error %v.", err)
	}
	return err
}

type ConsensusResult struct {
	Signatures [][]byte
	Result     error
}

// request msg
type RequestMessage struct {
	Request *Request
}

type Request struct {
	Timestamp int64
	Account   account.Account
	Payload   *types.Block
}

// proposal msg
type ProposalMessage struct {
	Proposal *Proposal
}

type Proposal struct {
	Account   account.Account
	Timestamp int64
	Payload   *types.Block
	Signature []byte
}

// response msg
type ResponseMessage struct {
	Response *Response
}

type Response struct {
	Account     account.Account
	Timestamp   int64
	Digest      types.Hash
	Signature   []byte
	BlockHeight uint64
}

// online request
type OnlineRequestMessage struct {
	OnlineRequest *OnlineRequest
}

type OnlineRequest struct {
	Account     account.Account
	Timestamp   int64
	BlockHeight uint64
}

// online response
type OnlineResponseMessage struct {
	OnlineResponse *OnlineResponse
}

type OnlineResponse struct {
	Account     account.Account
	Timestamp   int64
	BlockHeight uint64
	ViewNum     uint64
	Nodes       []account.Account
	Master      account.Account
}

// sync role assignment from other node
type SyncRoleAssignmentReqMessage struct {
	SyncRoleAssignmentReq *SyncRoleAssignmentReq
}

type SyncRoleAssignmentReq struct {
	Account   account.Account
	Timestamp int64
	ViewNum   uint64
}

// send sync role assignment to other node
type SyncRoleAssignmentRespMessage struct {
	SyncRoleAssignmentResp *SyncRoleAssignmentResp
}

type SyncRoleAssignmentResp struct {
	Account   account.Account
	ViewNum   uint64
	Master    account.Account
	Timestamp int64
}

// commit msg
type CommitMessage struct {
	Commit *Commit
}

type Commit struct {
	Account    account.Account
	Timestamp  int64
	BlockHash  types.Hash
	Digest     types.Hash
	Signatures [][]byte
	Result     bool
}

// sync block request msg
type SyncBlockReqMessage struct {
	SyncBlockReq *SyncBlockReq
}

type SyncBlockReq struct {
	Account    account.Account
	Timestamp  int64
	BlockStart uint64
	BlockEnd   uint64
}

// sync block response msg
type SyncBlockRespMessage struct {
	SyncBlockResp *SyncBlockResp
}

type SyncBlockResp struct {
	// TODO: add signatures
	Blocks []*types.Block
}

// change view request msg
type ViewChangeReqMessage struct {
	ViewChange *ViewChangeReq
}

type ViewChangeReq struct {
	Account   account.Account
	Nodes     []account.Account
	Timestamp int64
	ViewNum   uint64
}

// send msg to specified destination
func Unicast(account account.Account, msgPayload []byte, MessageType MessageType, digest types.Hash) error {
	log.Info("send msg [type %v, digest %x] to %d with url %s.", MessageType, digest, account.Extension.Id, account.Extension.Url)
	err := sendMsgByUrl(account.Extension.Url, msgPayload)
	if nil != err {
		log.Error("send msg [type %v and digest %x] to %d with url %s occurs error %v.",
			MessageType, digest, account.Extension.Id, account.Extension.Url, err)
	}
	return err
}

func BroadcastPeers(msgPayload []byte, MessageType MessageType, digest types.Hash, peers []account.Account) {
	for _, peer := range peers {
		log.Info("broadcast to %d by url %s with message type %v and digest %x.",
			peer.Extension.Id, peer.Extension.Url, MessageType, digest)
		err := sendMsgByUrl(peer.Extension.Url, msgPayload)
		if nil != err {
			log.Error("broadcast to %d by url %s with message type %v and digest %x occur error %v.",
				peer.Extension.Id, peer.Extension.Url, MessageType, digest, err)
		}
	}
}

func BroadcastPeersFilter(msgPayload []byte, MessageType MessageType, digest types.Hash, peers []account.Account, black account.Account) {
	for _, peer := range peers {
		if peer != black {
			log.Info("broadcast to %d by url %s with message type %v and digest %x.",
				peer.Extension.Id, peer.Extension.Url, MessageType, digest)
			err := sendMsgByUrl(peer.Extension.Url, msgPayload)
			if nil != err {
				log.Error("broadcast to %d by url %s with message type %v and digest %x occur error %v.",
					peer.Extension.Id, peer.Extension.Url, MessageType, digest, err)
			}
		}
	}
}
