package misc

import (
	"sync"
)

type Event struct {
	Name string
	Data interface{}
}

type EventListener func(Event)

type EventEmitter struct {
	listeners map[string][]EventListener
	mu        sync.RWMutex
}

func NewEventEmitter() *EventEmitter {
	return &EventEmitter{
		listeners: make(map[string][]EventListener),
	}
}

func (ee *EventEmitter) Subscribe(eventName string, listener EventListener) {
	ee.mu.Lock()
	defer ee.mu.Unlock()
	ee.listeners[eventName] = append(ee.listeners[eventName], listener)
}

// Publish invokes every listener for the event synchronously, in the caller's
// goroutine. The TUI handlers that publish all run on the tview event loop, and
// the listeners mutate widgets (tables, trees) — which tview requires to happen
// on that single loop. Dispatching on detached goroutines (the previous
// behaviour) raced with drawing, so listeners must run inline here.
//
// The listener slice is copied under the read lock and the lock released before
// any listener runs, so a listener is free to (un)subscribe without deadlocking.
func (ee *EventEmitter) Publish(event Event) {
	ee.mu.RLock()
	listeners := ee.listeners[event.Name]
	ee.mu.RUnlock()

	for _, listener := range listeners {
		listener(event)
	}
}

// PublishAndWait is retained for call sites that document "wait for completion";
// Publish is already synchronous, so it simply delegates.
func (ee *EventEmitter) PublishAndWait(event Event) {
	ee.Publish(event)
}
