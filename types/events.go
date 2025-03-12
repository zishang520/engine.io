package types

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
)

const (
	// Version current version number
	EventVersion = "0.0.3"
	// DefaultMaxListeners is the number of max listeners per event
	// default EventEmitters will print a warning if more than x listeners are
	// added to it. This is a useful default which helps finding memory leaks.
	// Defaults to 0, which means unlimited
	EventDefaultMaxListeners = 0
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
		fn  Listener
		ptr uintptr
	}

	eventEntry struct {
		mu        sync.RWMutex
		listeners []*listener
	}

	emmiter struct {
		maxListeners atomic.Uint32
		evtListeners Map[EventName, *eventEntry]
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
func NewEventEmitter() EventEmitter {
	emmiter := &emmiter{
		evtListeners: Map[EventName, *eventEntry]{},
	}
	emmiter.SetMaxListeners(EventDefaultMaxListeners)

	return emmiter
}

func (e *emmiter) SetMaxListeners(n uint) {
	e.maxListeners.Store(uint32(n))
}

func (e *emmiter) GetMaxListeners() uint {
	return uint(e.maxListeners.Load())
}

func (e *emmiter) addListeners(evt EventName, listeners []*listener) error {
	if len(listeners) == 0 {
		return nil
	}

	evtEntry, _ := e.evtListeners.LoadOrStore(evt, &eventEntry{})

	evtEntry.mu.Lock()
	defer evtEntry.mu.Unlock()

	if maxListeners := e.maxListeners.Load(); maxListeners > 0 && len(evtEntry.listeners) >= int(maxListeners) {
		return fmt.Errorf("(events) warning: possible EventEmitter memory leak detected. %d listeners added. Use emitter.SetMaxListeners(n int) to increase limit.", len(evtEntry.listeners))
	}

	evtEntry.listeners = append(evtEntry.listeners, listeners...)
	return nil
}

func (e *emmiter) AddListener(evt EventName, listeners ...Listener) error {
	if len(listeners) == 0 {
		return nil
	}

	events := make([]*listener, len(listeners))
	for i, event := range listeners {
		events[i] = &listener{fn: event, ptr: reflect.ValueOf(event).Pointer()}
	}

	return e.addListeners(evt, events)
}

// Alias: [AddListener]
func (e *emmiter) On(evt EventName, listeners ...Listener) error {
	return e.AddListener(evt, listeners...)
}

func (e *emmiter) Emit(evt EventName, data ...any) {
	evtEntry, ok := e.evtListeners.Load(evt)
	if !ok {
		return
	}

	evtEntry.mu.RLock()
	if len(evtEntry.listeners) == 0 {
		evtEntry.mu.RUnlock()
		return
	}

	listeners := make([]*listener, len(evtEntry.listeners))
	copy(listeners, evtEntry.listeners)
	evtEntry.mu.RUnlock()

	for _, event := range listeners {
		if event != nil {
			event.fn(data...)
		}
	}
}

func (e *emmiter) EventNames() []EventName {
	return e.evtListeners.Keys()
}

func (e *emmiter) ListenerCount(evt EventName) int {
	evtEntry, ok := e.evtListeners.Load(evt)
	if !ok {
		return 0
	}

	evtEntry.mu.RLock()
	defer evtEntry.mu.RUnlock()

	return len(evtEntry.listeners)
}

func (e *emmiter) Listeners(evt EventName) []Listener {
	evtEntry, ok := e.evtListeners.Load(evt)
	if !ok {
		return nil
	}

	evtEntry.mu.RLock()
	defer evtEntry.mu.RUnlock()

	listeners := make([]Listener, len(evtEntry.listeners))
	for i, l := range evtEntry.listeners {
		listeners[i] = l.fn
	}

	return listeners
}

type oneTimeListener struct {
	fired *sync.Once

	evt     EventName
	emitter *emmiter
	fn      Listener
}

func (l *oneTimeListener) execute(vals ...any) {
	l.fired.Do(func() {
		defer l.emitter.RemoveListener(l.evt, l.fn)
		l.fn(vals...)
	})
}

func (e *emmiter) Once(evt EventName, listeners ...Listener) error {
	if len(listeners) == 0 {
		return nil
	}

	events := make([]*listener, len(listeners))
	for i, event := range listeners {
		oneTime := &oneTimeListener{fired: &sync.Once{}, evt: evt, emitter: e, fn: event}
		events[i] = &listener{fn: oneTime.execute, ptr: reflect.ValueOf(event).Pointer()}
	}
	return e.addListeners(evt, events)
}

// RemoveListener removes the specified listener from the listener array for the event named eventName.
func (e *emmiter) RemoveListener(evt EventName, listener Listener) bool {
	if listener == nil {
		return false
	}

	evtEntry, ok := e.evtListeners.Load(evt)

	if !ok {
		return false
	}

	evtEntry.mu.Lock()
	defer evtEntry.mu.Unlock()

	if len(evtEntry.listeners) == 0 {
		return false
	}

	targetPtr := reflect.ValueOf(listener).Pointer()

	for i, listener := range evtEntry.listeners {
		if listener.ptr == targetPtr {
			evtEntry.listeners = append(evtEntry.listeners[:i], evtEntry.listeners[i+1:]...)
			return true
		}
	}
	return false
}

func (e *emmiter) RemoveAllListeners(evt EventName) bool {
	_, loaded := e.evtListeners.LoadAndDelete(evt)
	return loaded
}

func (e *emmiter) Clear() {
	e.evtListeners.Clear()
}

func (e *emmiter) Len() int {
	return e.evtListeners.Len()
}
