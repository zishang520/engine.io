// Source: https://github.com/kataras/go-events
// Package events provides simple EventEmitter support for Go Programming Language
package events

import (
	"github.com/zishang520/engine.io/v2/types"
)

const (
	// Version current version number
	Version = types.EventVersion
	// DefaultMaxListeners is the number of max listeners per event
	// default EventEmitters will print a warning if more than x listeners are
	// added to it. This is a useful default which helps finding memory leaks.
	// Defaults to 0, which means unlimited
	DefaultMaxListeners = types.EventDefaultMaxListeners
)

type (
	// EventName is just a type of string, it's the event name
	EventName = types.EventName
	// Listener is the type of a Listener, it's a func which receives any,optional, arguments from the caller/emmiter
	Listener = types.Listener
	// Events the type for registered listeners, it's just a map[string][]func(...any)
	Events = types.Events
	// EventEmitter is the message/or/event manager
	EventEmitter = types.EventEmitter
)

// New returns a new, empty, EventEmitter
func New() EventEmitter {
	return types.NewEventEmitter()
}
