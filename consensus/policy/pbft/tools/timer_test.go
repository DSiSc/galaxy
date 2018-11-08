package tools

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func MockManager() Manager {
	manager := NewManagerImpl()
	return manager
}

func TestNewTimerImpl(t *testing.T) {
	asserts := assert.New(t)
	mr := MockManager()
	timer := NewTimerImpl(mr)
	temp := timer.(*timerImpl)
	asserts.NotNil(temp.manager)
	asserts.Equal(mr, temp.manager)
}

func TestTimerImpl_Halt(t *testing.T) {
	asserts := assert.New(t)
	mr := MockManager()
	timer := NewTimerImpl(mr)
	temp := timer.(*timerImpl)
	temp.Halt()
	// if the channel is closed, isClose is false
	_, isClose := <-temp.threaded.exit
	asserts.Equal(false, isClose)
}

func TestTimerImpl_Reset(t *testing.T) {
	mr := MockManager()
	timer := NewTimerImpl(mr)
	temp := timer.(*timerImpl)
	temp.Reset(time.Duration(2), func() {})
}

func TestTimerImpl_SoftReset(t *testing.T) {
	mr := MockManager()
	timer := NewTimerImpl(mr)
	temp := timer.(*timerImpl)
	temp.SoftReset(time.Duration(2), func() {})
}

func TestTimerImpl_Stop(t *testing.T) {
	mr := MockManager()
	timer := NewTimerImpl(mr)
	time.Sleep(1 * time.Second)
	temp := timer.(*timerImpl)
	temp.Reset(time.Duration(1), func() {})
	time.Sleep(2 * time.Second)
	temp.Stop()
}
