package common

import "errors"

type Roler int

// Role
const (
	Master  Roler = iota // Master --> 0
	Slave                // Slave --> 1
	Normal               // Normal --> 2, node that not participation in consensus
	UnKnown              // UnKnown --> 3, node that nobody knows
)

const (
	SoloPolicy = "solo"
	DposPolicy = "dpos"
)

var (
	AssignmentNotBeExecute = errors.New("assignments role has not be executed")
)
