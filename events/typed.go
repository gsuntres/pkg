package events

import (
	"fmt"
	"strings"
	"sync"
)

type Event any

type Handler[T Event] func(*T) error

type TypedEventBus struct {
	mu sync.RWMutex
	subs map[string][]any
}

func NewTypedEventBus() *TypedEventBus {
	return &TypedEventBus{
		subs: make(map[string][]any),
	}
}

func typeName[T Event]() string {
	var t T
	v := fmt.Sprintf("%T", t)

	return strings.TrimLeft(v, "*")
}

func Count[T Event](bus *TypedEventBus) int {
	name := typeName[T]()

	bus.mu.RLock()
	subs := bus.subs[name]
	bus.mu.RUnlock()
	
	return len(subs)
}

func Subscribe[T Event](bus *TypedEventBus, h Handler[T]) (cancel func()) {
	name := typeName[T]()

	bus.mu.Lock()
	idx := len(bus.subs[name])
	bus.subs[name] = append(bus.subs[name], h)
	bus.mu.Unlock()
	
	return func() {
		idx := idx
		bus.mu.Lock()
		handlers := bus.subs[name]
		bus.subs[name] = append(handlers[:idx], handlers[idx+1:]...)
		bus.mu.Unlock()
	}
}

func Publish[T Event](bus *TypedEventBus, e *T) error {
	name := typeName[T]()

	bus.mu.RLock()
	hdlers := bus.subs[name]
	handlers := make([]Handler[T], 0, len(hdlers))
  for _, h := range hdlers {
    handlers = append(handlers, h.(Handler[T]))
  }
	bus.mu.RUnlock()

	ln := len(handlers)
	if ln > 0 {

		for _, h := range handlers {
			if err := h(e); err != nil {
				return err
			}
		}
	}

	return nil
}