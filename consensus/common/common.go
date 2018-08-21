package common

import (
	"github.com/DSiSc/producer/common"
)

type Version uint8

// Base proposal
type Proposal struct {
	Block *common.Block
}

type ConsensusStatus uint8

// Consensus status
const (
	Proposing ConsensusStatus = iota
	Propose
	Approve
	Reject
	Commited
)
