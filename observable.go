package observable

import (
	"reflect"
	"strings"
	"sync"
)

// ALL_EVENTS_NAMESPACE event key uset to listen and remove all the events
const ALL_EVENTS_NAMESPACE = "*"

// Structs

// private struct
type callback struct {
	fn        reflect.Value
	isUnique  bool
	isTyped   bool
	wasCalled bool
}

// Observable struct
type Observable struct {
	Callbacks map[string][]callback
	*sync.RWMutex
}

// Public API

// New - returns a new observable reference
func New() *Observable {
	return &Observable{
		make(map[string][]callback),
		&sync.RWMutex{},
	}
}

// On - adds a callback function
func (o *Observable) On(event string, cb interface{}) *Observable {
	fn := reflect.ValueOf(cb)
	events := strings.Fields(event)
	isTyped := len(events) > 1

	for _, s := range events {
		o.addCallback(s, fn, false, isTyped)
	}
	return o
}

// Trigger - a particular event passing custom arguments
func (o *Observable) Trigger(event string, params ...interface{}) *Observable {

	// get the arguments we want to pass to our listeners callbaks
	arguments := make([]reflect.Value, len(params))

	// get all the arguments
	for i, param := range params {
		arguments[i] = reflect.ValueOf(param)
	}
	// get all the list of events space separated
	events := strings.Fields(event)

	for _, s := range events {
		o.dispatchEvent(s, arguments)
	}
	// trigger the all events callback whenever this event was defined
	if o.hasEvent(ALL_EVENTS_NAMESPACE) && event != ALL_EVENTS_NAMESPACE {
		o.dispatchEvent(ALL_EVENTS_NAMESPACE, append([]reflect.Value{reflect.ValueOf(event)}, arguments...))
	}

	return o
}

// Off - stop listening a particular event
func (o *Observable) Off(event string, args ...interface{}) *Observable {
	if event == ALL_EVENTS_NAMESPACE {
		// wipe all the event listeners
		o.cleanEvent(event)
		return o
	}
	events := strings.Fields(event)
	for _, s := range events {
		if len(args) == 0 {
			o.cleanEvent(s)
		} else if len(args) == 1 {
			fn := reflect.ValueOf(args[0])
			o.removeEvent(s, fn)
		} else {
			panic("Multiple off callbacks are not supported")
		}
	}

	return o
}

// One - call the callback only once
func (o *Observable) One(event string, cb interface{}) *Observable {
	fn := reflect.ValueOf(cb)
	events := strings.Fields(event)
	isTyped := len(events) > 1

	for _, s := range events {
		o.addCallback(s, fn, true, isTyped)
	}
	return o
}
