package common

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/validator/tools/account"
)

type Version uint64

// Base proposal
type Proposal struct {
	Block     *types.Block
	Timestamp int64
}

// BFTRequest that with bft policy
type Request struct {
	Id      uint64
	Payload *Proposal
	Status  ConsensusStatus
}

type ConsensusStatus uint8

const (
	Proposing ConsensusStatus = iota // Proposing --> 0  prepare to launch a proposal
	Propose                          // Propose --> 1 propose for a proposal
	Approve                          // Approve --> 2 response of participate which accept the proposal
	Reject                           // Reject --> 3 response of participate which reject the proposal
	Committed                        // Committed --> 4 proposal has been accepted by participates with consensus policy
)

const (
	SoloPolicy       = "solo"
	SoloConsensusNum = uint8(1)
	BftPolicy        = "bft"
	FbftPolicy       = "fbft"
	DbftPolicy       = "dbft"
)

const (
	MaxBufferLen       = 1024 * 256
	DefaultViewNum     = uint64(0)
	DefaultWalterLevel = int64(1)
	DefaultBlockHeight = uint64(0)
)

var (
	ErrorsNewBlockChainByBlockHash = errors.New("get block chain by hash failed")
)

type ViewStatus string

const ViewChanging ViewStatus = "ViewChanging"
const ViewNormal ViewStatus = "ViewNormal"

type ViewRequestState string

const Viewing ViewRequestState = "Viewing"
const ViewEnd ViewRequestState = "ViewEnd"

type OnlineState string

const GoOnline OnlineState = "GoOnline"
const Online OnlineState = "Online"

type ConsensusResult struct {
	View        uint64
	Participate []account.Account
	Master      account.Account
}

type MessageSignal uint8

const (
	_ = MessageSignal(iota)
	ReceiveResponseSignal
)

func Sum(bz []byte) []byte {
	hash := sha256.Sum256(bz)
	return hash[:types.HashLength]
}

func HeaderHash(block *types.Block) (hash types.Hash) {
	var defaultHash types.Hash
	if !bytes.Equal(block.HeaderHash[:], defaultHash[:]) {
		copy(hash[:], block.HeaderHash[:])
		return
	}
	jsonByte, _ := json.Marshal(block.Header)
	sumByte := Sum(jsonByte)
	copy(hash[:], sumByte)
	return
}
