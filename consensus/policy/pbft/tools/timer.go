package tools

import (
	"github.com/DSiSc/craft/log"
	"time"
)

// Timer is an interface for managing time driven events
// the special contract Timer gives which a traditional golang
// timer does not, is that if the event thread calls stop, or reset
// then even if the timer has already fired, the event will not be
// delivered to the event queue
type Timer interface {
	SoftReset(duration time.Duration, event Event) // start a new countdown, only if one is not already started
	Reset(duration time.Duration, event Event)     // start a new countdown, clear any pending events
	Stop()                                         // stop the countdown, clear any pending events
	Halt()                                         // Stops the Timer thread
}

// TimerFactory abstracts the creation of Timers, as they may
// need to be mocked for testing
type TimerFactory interface {
	CreateTimer() Timer // Creates an Timer which is stopped
}

// timerStart is used to deliver the start request to the eventTimer thread
type timerStart struct {
	hard     bool          // Whether to reset the timer if it is running
	event    Event         // What event to push onto the event queue
	duration time.Duration // How long to wait before sending the event
}

// timerImpl is an implementation of Timer
type timerImpl struct {
	threaded                   // Gives us the exit chan
	timerChan <-chan time.Time // When non-nil, counts down to preparing to do the event
	startChan chan *timerStart // Channel to deliver the timer start events to the service go routine
	stopChan  chan struct{}    // Channel to deliver the timer stop events to the service go routine
	manager   Manager          // The event manager to deliver the event to after timer expiration
}

// newTimer creates a new instance of timerImpl
func NewTimerImpl(manager Manager) Timer {
	et := &timerImpl{
		startChan: make(chan *timerStart),
		stopChan:  make(chan struct{}),
		threaded:  threaded{make(chan struct{})},
		manager:   manager,
	}
	go et.loop()
	return et
}

// loop is where the timer thread lives, looping
func (et *timerImpl) loop() {
	var eventDestChan chan<- Event
	var event Event

	for {
		// A little state machine, relying on the fact that nil channels will block on read/write indefinitely
		select {
		case start := <-et.startChan:
			if et.timerChan != nil {
				if start.hard {
					log.Debug("Resetting a running timer")
				} else {
					continue
				}
			}
			log.Debug("Starting timer")
			et.timerChan = time.After(start.duration)
			if eventDestChan != nil {
				log.Debug("Timer cleared pending event")
			}
			event = start.event
			eventDestChan = nil
		case <-et.stopChan:
			if et.timerChan == nil && eventDestChan == nil {
				log.Debug("Attempting to stop an unfired idle timer")
			}
			et.timerChan = nil
			log.Debug("Stopping timer")
			if eventDestChan != nil {
				log.Debug("Timer cleared pending event")
			}
			eventDestChan = nil
			event = nil
		case <-et.timerChan:
			log.Debug("Event timer fired")
			et.timerChan = nil
			eventDestChan = et.manager.Queue()
		case eventDestChan <- event:
			log.Debug("Timer event delivered")
			eventDestChan = nil
		case <-et.exit:
			log.Debug("Halting timer")
			return
		}
	}
}

// softReset tells the timer to start a new countdown, only if it is not currently counting down
// this will not clear any pending events
func (et *timerImpl) SoftReset(timeout time.Duration, event Event) {
	et.startChan <- &timerStart{
		duration: timeout,
		event:    event,
		hard:     false,
	}
}

// reset tells the timer to start counting down from a new timeout, this also clears any pending events
func (et *timerImpl) Reset(timeout time.Duration, event Event) {
	et.startChan <- &timerStart{
		duration: timeout,
		event:    event,
		hard:     true,
	}
}

// stop tells the timer to stop, and not to deliver any pending events
func (et *timerImpl) Stop() {
	et.stopChan <- struct{}{}
}

func (et *timerImpl) Halt() {
	select {
	case <-et.threaded.exit:
		log.Warn("Attempted to halt a threaded object twice")
	default:
		close(et.threaded.exit)
	}
}
