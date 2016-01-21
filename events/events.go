package events

import (
	"github.com/samalba/dockerclient"
)

type (
	EventHandler struct {
		eventChan chan (*dockerclient.Event)
	}
)

func NewEventHandler(eventChan chan (*dockerclient.Event)) (*EventHandler, error) {
	h := &EventHandler{
		eventChan: eventChan,
	}

	return h, nil
}

func (h *EventHandler) Handle(e *dockerclient.Event, ec chan error, args ...interface{}) {
	go func() {
		h.eventChan <- e
	}()
}
