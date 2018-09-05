package common

import (
	"github.com/DSiSc/craft/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_Role(t *testing.T) {
	asserts := assert.New(t)
	asserts.Equal(0, int(Proposing))
	asserts.Equal(1, int(Propose))
	asserts.Equal(2, int(Approve))
	asserts.Equal(3, int(Reject))
	asserts.Equal(4, int(Committed))
}

var MockHash = types.Hash{
	0x1d, 0xcf, 0x7, 0xba, 0xfc, 0x42, 0xb0, 0x8d, 0xfd, 0x23, 0x9c, 0x45, 0xa4, 0xb9, 0x38, 0xd,
	0x8d, 0xfe, 0x5d, 0x6f, 0xa7, 0xdb, 0xd5, 0x50, 0xc9, 0x25, 0xb1, 0xb3, 0x4, 0xdc, 0xc5, 0x1c,
}

var MockHeaderHash = types.Hash{
	0xcc, 0x88, 0x1c, 0x28, 0x30, 0x38, 0x50, 0x46, 0x2c, 0xcb, 0xae, 0xe5, 0xa4, 0x88, 0x85, 0x75,
	0xdf, 0xae, 0xd7, 0xd3, 0x39, 0x17, 0x9a, 0xfc, 0x9c, 0x4, 0x5e, 0xcd, 0x98, 0x8a, 0x39, 0xdd,
}

func MockBlock() *types.Block {
	return &types.Block{
		Header: &types.Header{
			ChainID:       1,
			PrevBlockHash: MockHash,
			StateRoot:     MockHash,
			TxRoot:        MockHash,
			ReceiptsRoot:  MockHash,
			Height:        1,
			Timestamp:     uint64(time.Date(2018, time.August, 28, 0, 0, 0, 0, time.UTC).Unix()),
			MixDigest:     MockHash,
		},
		Transactions: make([]*types.Transaction, 0),
	}
}

func TestBlockHash(t *testing.T) {
	block := MockBlock()
	headerHash := HeaderHash(block)
	assert.Equal(t, MockHeaderHash, headerHash)
	headerHash = HeaderHash(block)
	assert.Equal(t, MockHeaderHash, headerHash)
}
