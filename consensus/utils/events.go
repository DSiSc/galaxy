package utils

// Event is a type meant to clearly convey that the return type or parameter to a function will be supplied to/from an events.Manager
type Event interface{}

// Receiver is a consumer of events, ProcessEvent will be called serially
// as events arrive
type Receiver interface {
	// ProcessEvent delivers an event to the Receiver, if it returns non-nil, the return is the next processed event
	ProcessEvent(e Event) Event
}

// SendEvent performs the event loop on a receiver to completion
func SendEvent(receiver Receiver, event Event) {
	next := event
	for {
		// If an event returns something non-nil, then process it as a new event
		next = receiver.ProcessEvent(next)
		if next == nil {
			break
		}
	}
}
