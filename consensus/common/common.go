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

const (
	Proposing ConsensusStatus = iota // Proposing --> 0  prepare to launch a proposal
	Propose                          // Propose --> 1 propose for a proposal
	Approve                          // Approve --> 2 response of participate which accept the proposal
	Reject                           // Reject --> 3 response of participate which reject the proposal
	Commited                         // Commited --> 4 the proposal has be accepted by all participate for some consensus policy
)
