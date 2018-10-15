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
	Block *types.Block
}

type ConsensusStatus uint8

const (
	Proposing ConsensusStatus = iota // Proposing --> 0  prepare to launch a proposal
	Propose                          // Propose --> 1 propose for a proposal
	Approve                          // Approve --> 2 response of participate which accept the proposal
	Reject                           // Reject --> 3 response of participate which reject the proposal
	Committed                        // Committed --> 4 proposal has been accepted by participates with consensus policy
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
