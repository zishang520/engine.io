package events

import (
	"fmt"
	"testing"
	"time"
)

var _event = New()

var testEvents = Events{
	"user_created": []Listener{
		func(payload ...any) {
			fmt.Printf("A new User just created!\n")
		},
		func(payload ...any) {
			fmt.Printf("A new User just created, *from second event listener\n")
		},
	},
	"user_joined": []Listener{func(payload ...any) {
		user := payload[0].(string)
		room := payload[1].(string)
		fmt.Printf("%s joined to room: %s\n", user, room)
	}},
	"user_left": []Listener{func(payload ...any) {
		user := payload[0].(string)
		room := payload[1].(string)
		fmt.Printf("%s left from the room: %s\n", user, room)
	}},
}

func createUser(user string) {
	_event.Emit("user_created", user)
}

func joinUserTo(user string, room string) {
	_event.Emit("user_joined", user, room)
}

func leaveFromRoom(user string, room string) {
	_event.Emit("user_left", user, room)
}

func ExampleEvents() {
	// regiter our events to the default event emmiter
	for evt, listeners := range testEvents {
		_event.On(evt, listeners...)
	}

	user := "user1"
	room := "room1"

	createUser(user)
	joinUserTo(user, room)
	leaveFromRoom(user, room)

	// Output:
	// A new User just created!
	// A new User just created, *from second event listener
	// user1 joined to room: room1
	// user1 left from the room: room1
}

func TestEvents(t *testing.T) {
	e := New()
	expectedPayload := "this is my payload"

	e.On("my_event", func(payload ...any) {
		if len(payload) <= 0 {
			t.Fatal("Expected payload but got nothing")
		}

		if s, ok := payload[0].(string); !ok {
			t.Fatalf("Payload is not the correct type, got: %#v", payload[0])
		} else if s != expectedPayload {
			t.Fatalf("Eexpected %s, got: %s", expectedPayload, s)
		}
	})

	e.Emit("my_event", expectedPayload)
	if e.Len() != 1 {
		t.Fatalf("Length of the events is: %d, while expecting: %d", e.Len(), 1)
	}

	if e.Len() != 1 {
		t.Fatalf("Length of the listeners is: %d, while expecting: %d", e.ListenerCount("my_event"), 1)
	}

	e.RemoveAllListeners("my_event")
	if e.Len() != 0 {
		t.Fatalf("Length of the events is: %d, while expecting: %d", e.Len(), 0)
	}

	if e.Len() != 0 {
		t.Fatalf("Length of the listeners is: %d, while expecting: %d", e.ListenerCount("my_event"), 0)
	}
}

func TestEventsOnce(t *testing.T) {
	// on default
	_event.Clear()

	var count = 0
	_event.Once("my_event", func(payload ...any) {
		if count > 0 {
			t.Fatalf("Once's listener fired more than one time! count: %d", count)
		}
		if l := len(payload); l != 2 {
			t.Fatalf("Once's listeners (from Listeners) should be: %d but has: %d", 2, l)
		}
		count++
	})
	if l := _event.ListenerCount("my_event"); l != 1 {
		t.Fatalf("Real  event's listeners should be: %d but has: %d", 1, l)
	}

	if l := len(_event.Listeners("my_event")); l != 1 {
		t.Fatalf("Real  event's listeners (from Listeners) should be: %d but has: %d", 1, l)
	}

	for i := 0; i < 10; i++ {
		_event.Emit("my_event", "foo", "foo1")
	}

	time.Sleep(10 * time.Millisecond)

	if l := _event.ListenerCount("my_event"); l > 0 {
		t.Fatalf("Real event's listeners length count should be: %d but has: %d", 0, l)
	}

	if l := len(_event.Listeners("my_event")); l > 0 {
		t.Fatalf("Real event's listeners length count ( from Listeners) should be: %d but has: %d", 0, l)
	}

}

func TestRemoveListener(t *testing.T) {
	// on default
	e := New()

	var count = 0
	listener := func(payload ...any) {
		if count > 1 {
			t.Fatal("Event listener should be removed")
		}

		count++
	}

	once := func(payload ...any) {}

	e.Once("once_event", once)
	e.AddListener("my_event", listener)
	e.AddListener("my_event", func(payload ...any) {})
	e.AddListener("another_event", func(payload ...any) {})

	e.Emit("my_event")

	if e.RemoveListener("once_event", once) != true {
		t.Fatal("Should return 'true' when removes found once listener")
	}

	if e.ListenerCount("once_event") != 0 {
		t.Fatal("Length of 'once_event' event listeners must be 0")
	}

	if e.RemoveListener("my_event", listener) != true {
		t.Fatal("Should return 'true' when removes found listener")
	}

	if e.RemoveListener("foo_bar", listener) != false {
		t.Fatal("Should return 'false' when removes nothing")
	}

	if e.Len() != 3 {
		t.Fatal("Length of all listeners must be 2")
	}

	if e.ListenerCount("my_event") != 1 {
		t.Fatal("Length of 'my_event' event listeners must be 1")
	}

	e.Emit("my_event")
}
