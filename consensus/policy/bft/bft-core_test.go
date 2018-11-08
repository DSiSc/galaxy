package bft

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var id uint64 = 0

func TestNewBFTCore(t *testing.T) {
	receive := NewBFTCore(id)
	assert.NotNil(t, receive)
	bftcore := receive.(*bftcore)
	assert.Equal(t, id, bftcore.id)
}

func TestBftcore_ProcessEvent(t *testing.T) {
	receive := NewBFTCore(id)
	assert.NotNil(t, receive)
	bftcore := receive.(*bftcore)
	err := bftcore.ProcessEvent(nil)
	assert.Nil(t, err)
}
