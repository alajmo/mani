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

func (ee *EventEmitter) Publish(event Event) {
	ee.mu.RLock()
	defer ee.mu.RUnlock()
	if listeners, ok := ee.listeners[event.Name]; ok {
		for _, listener := range listeners {
			go listener(event)
		}
	}
}

func (ee *EventEmitter) PublishAndWait(event Event) {
	ee.mu.RLock()
	listeners := ee.listeners[event.Name]
	ee.mu.RUnlock()

	var wg sync.WaitGroup
	for _, listener := range listeners {
		wg.Add(1)
		go func(l EventListener) {
			defer wg.Done()
			l(event)
		}(listener)
	}
	wg.Wait()
}
