package bft

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockEvent struct {
	processed bool
}

func (self *mockEvent) ProcessEvent(e Event) Event {
	self.processed = true
	return nil
}

func newMockEvent() Receiver {
	return &mockEvent{
		processed: false,
	}
}

func TestSendEvent(t *testing.T) {
	mock := newMockEvent()
	event := mock.(*mockEvent)
	assert.NotNil(t, event)
	assert.Equal(t, false, event.processed)

	mock.ProcessEvent(nil)
	assert.Equal(t, true, event.processed)
}
