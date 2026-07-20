package events

import (
	"testing"
)

type MyEvent struct {
	Name string
}

func TestDefault(t *testing.T) {
	bus := NewTypedEventBus()
	if Count[MyEvent](bus) != 0 {
		t.Fatalf("Should have no subscribers by default")
	}
}

func TestOneSub(t *testing.T) {
	bus := NewTypedEventBus()

	Subscribe(bus, func(e *MyEvent) error { return nil })

	if Count[MyEvent](bus) != 1 {
		t.Fatalf("Should have one subscriber")
	}
}