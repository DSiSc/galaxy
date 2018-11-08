package bft

import (
	"github.com/DSiSc/galaxy/consensus/policy/bft/messages"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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

	var mock_request = messages.Request{
		Timestamp: time.Now().Unix(),
		Payload:   nil,
	}

	err = bftcore.ProcessEvent(mock_request)
	assert.Nil(t, err)

	var mock_proposal = messages.Propoasl{
		Timestamp: time.Now().Unix(),
		Payload:   nil,
	}
	err = bftcore.ProcessEvent(mock_proposal)
	assert.Nil(t, err)
}
