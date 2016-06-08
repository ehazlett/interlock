package events

import (
	"testing"

	etypes "github.com/docker/engine-api/types/events"
)

func TestEventHandler(t *testing.T) {
	tChan := make(chan *etypes.Message)

	errChan := make(chan error)

	go func() {
		for err := range errChan {
			t.Fatal(err)
		}
	}()

	h, err := NewEventHandler(tChan)
	if err != nil {
		t.Fatal(err)
	}

	testEvent := &etypes.Message{
		Type: "testevent",
	}

	go h.Handle(testEvent, errChan, nil)

	v := <-tChan

	if v.Type != "testevent" {
		t.Fatalf("unexpected event type %s", v.Type)
	}
}
