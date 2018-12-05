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

//MsgType is the message type
type MsgType uint32

const (
	NIL = MsgType(iota)
	RequestMsgType
	ProposalMsgType
	ResponseMsgType
	CommitMsgType
	SyncBlockReqMsgType
	SyncBlockRespMsgType
	ViewChangeMsgReqType
)

type Msg struct {
	MsgType MsgType
	PayLoad interface{}
}

type MsgHeader struct {
	Magic   uint32
	MsgType MsgType
	Length  uint32
}

// EncodeMessage encode message to byte array.
func EncodeMessage(msg Msg) ([]byte, error) {
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
func buildMessageHeader(msg Msg, len int) (*MsgHeader, error) {
	header := &MsgHeader{
		Magic:   0,
		MsgType: msg.MsgType,
		Length:  uint32(len),
	}
	return header, nil
}

// encodeMessageHeader encode message header to byte array.
func encodeMessageHeader(header *MsgHeader) ([]byte, error) {
	buf := make([]byte, 12)
	binary.LittleEndian.PutUint32(buf, header.Magic)
	binary.LittleEndian.PutUint32(buf[4:], uint32(header.MsgType))
	binary.LittleEndian.PutUint32(buf[8:], header.Length)
	return buf, nil
}

// ReadMessage read message
func ReadMessage(reader io.Reader) (Msg, error) {
	header, err := readMessageHeader(reader)
	if err != nil {
		return Msg{}, err
	}

	body := make([]byte, header.Length)
	_, err = io.ReadFull(reader, body)
	if err != nil {
		return Msg{}, err
	}

	return decodeMessage(header.MsgType, body)
}

// read message header from reader.
func readMessageHeader(reader io.Reader) (MsgHeader, error) {
	header := MsgHeader{}
	err := binary.Read(reader, binary.LittleEndian, &header)
	return header, err
}

// make empty message according to the message type
func makeEmptyMessage(msgType MsgType) (interface{}, error) {
	switch msgType {
	case RequestMsgType:
		return &RequestMessage{}, nil
	case ProposalMsgType:
		return &ProposalMessage{}, nil
	case ResponseMsgType:
		return &ResponseMessage{}, nil
	case CommitMsgType:
		return &CommitMessage{}, nil
	case SyncBlockReqMsgType:
		return &SyncBlockReqMessage{}, nil
	case SyncBlockRespMsgType:
		return &SyncBlockRespMessage{}, nil
	case ViewChangeMsgReqType:
		return &ViewChangeReqMessage{}, nil
	default:
		return nil, fmt.Errorf("unknown message type %v", msgType)
	}
}

func decodeMessage(msgType MsgType, rawMsg []byte) (Msg, error) {
	payload, err := makeEmptyMessage(msgType)
	if nil != err {
		return Msg{}, err
	}
	err = json.Unmarshal(rawMsg, payload)
	if nil != err {
		log.Error("unmarshal rawMsg failed with err %v.", err)
		return Msg{}, err
	}
	return Msg{MsgType: msgType, PayLoad: payload}, nil
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
