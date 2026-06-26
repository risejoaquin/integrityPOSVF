package module

import (
	"log"
	"sync"
)

type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan interface{}
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan interface{}),
	}
}

func (b *EventBus) Subscribe(event string) <-chan interface{} {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan interface{}, 100)
	b.subscribers[event] = append(b.subscribers[event], ch)
	return ch
}

func (b *EventBus) Publish(event string, data interface{}) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if chans, ok := b.subscribers[event]; ok {
		for _, ch := range chans {
			select {
			case ch <- data:
				// Successfully sent
			default:
				log.Printf("Warning: channel for event %s is full, dropping event", event)
			}
		}
	}
}
