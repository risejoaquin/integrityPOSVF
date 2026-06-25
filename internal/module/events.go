package module

import "sync"

type EventBus struct {
	mu        sync.RWMutex
	callbacks map[string][]func(interface{})
}

func NewEventBus() *EventBus {
	return &EventBus{
		callbacks: make(map[string][]func(interface{})),
	}
}

func (b *EventBus) Subscribe(event string, callback func(interface{})) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.callbacks[event] = append(b.callbacks[event], callback)
}

func (b *EventBus) Publish(event string, data interface{}) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if callbacks, ok := b.callbacks[event]; ok {
		for _, cb := range callbacks {
			go cb(data)
		}
	}
}
