package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockReceive struct {
	init bool
}

func NewMockReceive() Receiver {
	return &MockReceive{
		init: true,
	}
}

var signal interface{}

func (instance *MockReceive) ProcessEvent(e Event) Event {
	signal = e
	return nil
}

func TestSendEvent2(t *testing.T) {
	receive := NewMockReceive()
	assert.Equal(t, true, receive.(*MockReceive).init)
	SendEvent(receive, 2)
	assert.Equal(t, 2, signal)
}
