// Source: https://github.com/kataras/go-events
// Package events provides simple EventEmitter support for Go Programming Language
package events

import (
	"fmt"
	"reflect"
	"sync"
)

const (
	// Version current version number
	Version = "0.0.3"
	// DefaultMaxListeners is the number of max listeners per event
	// default EventEmitters will print a warning if more than x listeners are
	// added to it. This is a useful default which helps finding memory leaks.
	// Defaults to 0, which means unlimited
	DefaultMaxListeners = 0
)

type (
	// EventName is just a type of string, it's the event name
	EventName string
	// Listener is the type of a Listener, it's a func which receives any,optional, arguments from the caller/emmiter
	Listener func(...any)
	// Events the type for registered listeners, it's just a map[string][]func(...any)
	Events map[EventName][]Listener
	// EventEmitter is the message/or/event manager
	EventEmitter interface {
		// AddListener is an alias for .On(eventName, listener).
		AddListener(EventName, ...Listener) error
		// Emit fires a particular event,
		// Synchronously calls each of the listeners registered for the event named
		// eventName, in the order they were registered,
		// passing the supplied arguments to each.
		Emit(EventName, ...any)
		// EventNames returns an array listing the events for which the emitter has registered listeners.
		// The values in the array will be strings.
		EventNames() []EventName
		// GetMaxListeners returns the max listeners for this emmiter
		// see SetMaxListeners
		GetMaxListeners() uint
		// ListenerCount returns the length of all registered listeners to a particular event
		ListenerCount(EventName) int
		// Listeners returns a copy of the array of listeners for the event named eventName.
		Listeners(EventName) []Listener
		// On registers a particular listener for an event, func receiver parameter(s) is/are optional
		On(EventName, ...Listener) error
		// Once adds a one time listener function for the event named eventName.
		// The next time eventName is triggered, this listener is removed and then invoked.
		Once(EventName, ...Listener) error
		// RemoveAllListeners removes all listeners, or those of the specified eventName.
		// Note that it will remove the event itself.
		// Returns an indicator if event and listeners were found before the remove.
		RemoveAllListeners(EventName) bool
		// RemoveListener removes given listener from the event named eventName.
		// Returns an indicator whether listener was removed
		RemoveListener(EventName, Listener) bool
		// Clear removes all events and all listeners, restores Events to an empty value
		Clear()
		// SetMaxListeners obviously this function allows the MaxListeners
		// to be decrease or increase. Set to zero for unlimited
		SetMaxListeners(uint)
		// Len returns the length of all registered events
		Len() int
	}

	listener struct {
		listener Listener
		ptr      uintptr
	}

	events map[EventName][]*listener

	emmiter struct {
		mu           sync.RWMutex
		maxListeners uint
		evtListeners events
	}
)

// CopyTo copies the event listeners to an EventEmitter
func (e Events) CopyTo(emitter EventEmitter) {
	if e != nil && len(e) > 0 {
		// register the events to/with their listeners
		for evt, listeners := range e {
			if len(listeners) > 0 {
				emitter.AddListener(evt, listeners...)
			}
		}
	}
}

// New returns a new, empty, EventEmitter
func New() EventEmitter {
	return &emmiter{maxListeners: DefaultMaxListeners, evtListeners: events{}}
}

func (e *emmiter) addListeners(evt EventName, listeners []*listener) error {
	if len(listeners) == 0 {
		return nil
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.evtListeners == nil {
		e.evtListeners = events{}
	}

	evts := e.evtListeners[evt]

	if e.maxListeners > 0 && len(evts) >= int(e.maxListeners) {
		return fmt.Errorf("(events) warning: possible EventEmitter memory leak detected. %d listeners added. Use emitter.SetMaxListeners(n int) to increase limit.", len(evts))
	}

	e.evtListeners[evt] = append(evts, listeners...)
	return nil
}

func (e *emmiter) AddListener(evt EventName, listeners ...Listener) error {
	if len(listeners) == 0 {
		return nil
	}
	events := make([]*listener, len(listeners))
	for i, event := range listeners {
		events[i] = &listener{listener: event, ptr: reflect.ValueOf(event).Pointer()}
	}
	return e.addListeners(evt, events)
}

func (e *emmiter) Emit(evt EventName, data ...any) {
	e.mu.RLock()
	listeners, ok := e.evtListeners[evt]
	if !ok || len(listeners) == 0 {
		e.mu.RUnlock()
		return
	}

	listenersCopy := make([]*listener, len(listeners))
	copy(listenersCopy, listeners)
	e.mu.RUnlock()

	for _, event := range listenersCopy {
		if event != nil {
			event.listener(data...)
		}
	}
}

func (e *emmiter) EventNames() (names []EventName) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.evtListeners == nil {
		return nil
	}

	for k := range e.evtListeners {
		names = append(names, k)
	}
	return names
}

func (e *emmiter) GetMaxListeners() uint {
	return e.maxListeners
}

func (e *emmiter) ListenerCount(evt EventName) int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.evtListeners == nil {
		return 0
	}
	return len(e.evtListeners[evt])
}

func (e *emmiter) Listeners(evt EventName) (listeners []Listener) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.evtListeners == nil {
		return nil
	}

	// do not pass any inactive/removed listeners(nil)
	for _, event := range e.evtListeners[evt] {
		if event != nil {
			listeners = append(listeners, event.listener)
		}
	}
	return listeners
}

func (e *emmiter) On(evt EventName, listeners ...Listener) error {
	return e.AddListener(evt, listeners...)
}

type oneTimeListener struct {
	fired *sync.Once

	evt      EventName
	emitter  *emmiter
	listener Listener
}

func (l *oneTimeListener) execute(vals ...any) {
	l.fired.Do(func() {
		defer l.emitter.RemoveListener(l.evt, l.listener)
		l.listener(vals...)
	})
}

func (e *emmiter) Once(evt EventName, listeners ...Listener) error {
	if len(listeners) == 0 {
		return nil
	}

	events := make([]*listener, len(listeners))
	for i, event := range listeners {
		oneTime := &oneTimeListener{fired: &sync.Once{}, evt: evt, emitter: e, listener: event}
		events[i] = &listener{listener: oneTime.execute, ptr: reflect.ValueOf(event).Pointer()}
	}
	return e.addListeners(evt, events)
}

func (e *emmiter) RemoveAllListeners(evt EventName) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.evtListeners == nil {
		return false // has nothing to remove
	}

	if _, ok := e.evtListeners[evt]; ok {
		delete(e.evtListeners, evt)
		return true
	}
	return false
}

// RemoveListener removes the specified listener from the listener array for the event named eventName.
func (e *emmiter) RemoveListener(evt EventName, listener Listener) bool {
	if listener == nil {
		return false
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.evtListeners == nil {
		return false
	}

	listeners, ok := e.evtListeners[evt]
	if !ok || len(listeners) == 0 {
		return false
	}

	listenerPointer := reflect.ValueOf(listener).Pointer()
	for i, event := range listeners {
		if event.ptr == listenerPointer {
			copy(listeners[i:], listeners[i+1:])
			e.evtListeners[evt] = listeners[:len(listeners)-1]
			return true
		}
	}
	return false
}

func (e *emmiter) Clear() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.evtListeners = events{}
}

func (e *emmiter) SetMaxListeners(n uint) {
	e.maxListeners = n
}

func (e *emmiter) Len() int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.evtListeners == nil {
		return 0
	}
	return len(e.evtListeners)
}
