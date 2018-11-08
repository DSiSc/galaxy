package tools

import (
	"fmt"
	"github.com/DSiSc/craft/log"
)

// threaded holds an exit channel to allow threads to break from a select
type threaded struct {
	exit chan struct{}
}

// Event is a type meant to clearly convey that the return type or parameter to
// a function will be supplied to/from an events.Manager
type Event interface{}

// Receiver is a consumer of events, ProcessEvent will be called serially
// as events arrive
type Receiver interface {
	// ProcessEvent delivers an event to the Receiver, if it returns non-nil, the return is the next processed event
	ProcessEvent(e Event) Event
}

// Manager provides a serialized interface for submitting events to
// a Receiver on the other side of the queue
type Manager interface {
	Inject(Event)         // A temporary interface to allow the event manager thread to skip the queue
	Queue() chan<- Event  // Get a write-only reference to the queue, to submit events
	SetReceiver(Receiver) // Set the target to route events to
	Start()               // Starts the Manager thread TODO, these thread management things should probably go away
	Halt()                // Stops the Manager thread
}

// managerImpl is an implementation of Manger
type managerImpl struct {
	threaded
	receiver Receiver
	events   chan Event
}

// NewManagerImpl creates an instance of managerImpl
func NewManagerImpl() Manager {
	return &managerImpl{
		events:   make(chan Event),
		threaded: threaded{make(chan struct{})},
	}
}

// Inject can only safely be called by the managerImpl thread itself, it skips the queue
func (em *managerImpl) Inject(event Event) {
	if em.receiver != nil {
		SendEvent(em.receiver, event)
	}
}

// SendEvent performs the event loop on a receiver to completion
func SendEvent(receiver Receiver, event Event) {
	next := event
	for {
		// If an event returns something non-nil, then process it as a new event
		next = receiver.ProcessEvent(next)
		if next == nil {
			fmt.Println("Ending.")
			break
		}
	}
}

// queue returns a write only reference to the event queue
func (em *managerImpl) Queue() chan<- Event {
	return em.events
}

// SetReceiver sets the destination for events
func (em *managerImpl) SetReceiver(receiver Receiver) {
	em.receiver = receiver
}

// Start creates the go routine necessary to deliver events
func (em *managerImpl) Start() {
	go em.eventLoop()
}

func (em *managerImpl) Halt() {
	select {
	case <-em.threaded.exit:
		log.Warn("Attempted to halt a threaded object twice")
	default:
		close(em.threaded.exit)
	}
}

// eventLoop is where the event thread loops, delivering events
func (em *managerImpl) eventLoop() {
	for {
		select {
		case next := <-em.events:
			em.Inject(next)
		case <-em.exit:
			log.Debug("eventLoop told to exit")
			return
		}
	}
}

// TimerFactoryImpl implements the TimerFactory
type timerFactoryImpl struct {
	manager Manager // The Manager to use in constructing the event timers
}

// NewTimerFactoryImpl creates a new TimerFactory for the given Manager
func NewTimerFactoryImpl(manager Manager) TimerFactory {
	return &timerFactoryImpl{manager: manager}
}

// CreateTimer creates a new timer which deliver events to the Manager for this factory
func (etf *timerFactoryImpl) CreateTimer() Timer {
	return newTimerImpl(etf.manager)
}

// newTimer creates a new instance of timerImpl
func newTimerImpl(manager Manager) Timer {
	et := &timerImpl{
		startChan: make(chan *timerStart),
		stopChan:  make(chan struct{}),
		threaded:  threaded{make(chan struct{})},
		manager:   manager,
	}
	go et.loop()
	return et
}
