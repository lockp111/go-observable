package observable

import (
	"reflect"
	"strings"
)

// Helpers

// Add a callback under a certain event namespace
func (o *Observable) addCallback(event string, cb interface{}, isUnique bool) *Observable {
	fn := reflect.ValueOf(cb)
	events := strings.Fields(event)
	isTyped := len(events) > 1

	for _, s := range events {
		go func() {
			// lock the struct
			o.Lock()
			defer o.Unlock()
			// does this namespace already exist?
			if !o.hasEvent(s) {
				o.Callbacks[s] = make([]callback, 1)
				o.Callbacks[s][0] = callback{fn, isUnique, isTyped, false}
			} else if !isUnique {
				o.Callbacks[s] = append(o.Callbacks[s], callback{fn, isUnique, isTyped, false})
			}
		}()
	}

	return o
}

// remove the events bound to the callback
func (o *Observable) removeEvent(event string, fn interface{}) {
	events := strings.Fields(event)
	// try to get the value of the function we want unsubscribe
	var n reflect.Value
	if fn != nil {
		n = reflect.ValueOf(fn)
	}

	for _, s := range events {

		if s == ALL_EVENTS_NAMESPACE {
			// lock the struct
			o.Lock()
			// wipe all the event listeners
			o.Callbacks = make(map[string][]callback)
			o.Unlock()
			return
		}
		go func() {
			// lock the struct
			o.Lock()
			defer o.Unlock()
			if o.hasEvent(s) && fn != nil {
				// loop all the callbacks registered under the event namespace
				for i, cb := range o.Callbacks[s] {
					if n == cb.fn {
						o.Callbacks[s] = append(o.Callbacks[s][:i], o.Callbacks[s][i+1:]...)
					}
				}
				// if there are no more callbacks using this namespace
				// delete the key from the map
				if len(o.Callbacks[s]) == 0 {
					delete(o.Callbacks, s)
				}
			}
		}()
	}

}

// dispatch the events using custom arguments
func (o *Observable) dispatchEvent(event string, arguments []reflect.Value) *Observable {
	// get all the list of events space separated
	events := strings.Fields(event)

	for _, s := range events {
		go func() {
			// lock the struct
			o.RLock()
			defer o.RUnlock()

			// check if the observable has already created this events map
			if o.hasEvent(s) {

				// loop all the callbacks
				// avoiding to call twice the ones registered with Observable.One
				for i, cb := range o.Callbacks[s] {

					if !cb.isUnique || cb.isUnique && !cb.wasCalled {
						// if the callback was registered with multiple events
						// we prepend the event namespace to the function arguments
						if cb.isTyped {
							cb.fn.Call(append([]reflect.Value{reflect.ValueOf(s)}, arguments...))
						} else {
							cb.fn.Call(arguments)
						}
					}

					o.Callbacks[s][i].wasCalled = true
				}
			}
		}()
	}

	return o
}

// check whether the Observable struct has already registered the event namespace
func (o *Observable) hasEvent(event string) bool {
	_, ok := o.Callbacks[event]
	return ok
}
