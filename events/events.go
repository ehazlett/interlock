package events

import (
	"github.com/docker/engine-api/types/events"
)

type (
	EventHandler struct {
		eventChan chan (*events.Message)
	}
)

func NewEventHandler(eventChan chan (*events.Message)) (*EventHandler, error) {
	h := &EventHandler{
		eventChan: eventChan,
	}

	return h, nil
}

func (h *EventHandler) Handle(e *events.Message, ec chan error, args ...interface{}) {
	h.eventChan <- e
}
