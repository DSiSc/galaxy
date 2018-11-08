package messages

import (
	"github.com/DSiSc/craft/types"
)

type Request struct {
	Timestamp int64
	Payload   *types.Block
}

type Propoasl struct {
	Timestamp int64
	Payload   *types.Block
}
