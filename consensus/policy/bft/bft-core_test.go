package bft

import (
	"github.com/DSiSc/galaxy/consensus/policy/bft/messages"
	"github.com/DSiSc/validator/tools/account"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var id uint64 = 0

func TestNewBFTCore(t *testing.T) {
	receive := NewBFTCore(id, true, nil)
	assert.NotNil(t, receive)
	bftcore := receive.(*bftCore)
	assert.Equal(t, id, bftcore.id)
}

func TestBftcore_ProcessEvent(t *testing.T) {
	receive := NewBFTCore(id, true, nil)
	assert.NotNil(t, receive)
	bftcore := receive.(*bftCore)
	err := bftcore.ProcessEvent(nil)
	assert.Nil(t, err)

	var mock_request = messages.Request{
		Timestamp: time.Now().Unix(),
		Payload:   nil,
	}

	err = bftcore.ProcessEvent(mock_request)
	assert.Nil(t, err)

	var mock_proposal = messages.Proposal{
		Timestamp: time.Now().Unix(),
		Payload:   nil,
	}
	err = bftcore.ProcessEvent(mock_proposal)
	assert.Nil(t, err)
}

func TestBftcore_Start(t *testing.T) {
	receive := NewBFTCore(id, true, nil)
	assert.NotNil(t, receive)
	bftcore := receive.(*bftCore)
	var account = account.Account{
		Extension: account.AccountExtension{
			Url: "127.0.0.1:8080",
		},
	}
	go bftcore.Start(account)
	time.Sleep(2 * time.Second)
}
