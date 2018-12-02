package common

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"github.com/DSiSc/craft/types"
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
	SOLO_POLICY        = "solo"
	SOLO_CONSENSUS_NUM = uint8(1)
	BFT_POLICY         = "bft"
	FBFT_POLICY        = "fbft"
	DBFT_POLICY        = "dbft"
)

type ViewStatus string

const ViewChanging ViewStatus = "ViewChanging"
const ViewNormal ViewStatus = "ViewNormal"

type ViewState string

const Viewing ViewState = "Viewing"
const ViewEnd ViewState = "ViewEnd"

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
