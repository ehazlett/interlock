package events

import (
	etypes "github.com/docker/docker/api/types/events"
)

type (
	EventHandler struct {
		eventChan chan *Message
	}
	Message struct {
		etypes.Message
	}
)

func NewEventHandler(eventChan chan *Message) (*EventHandler, error) {
	h := &EventHandler{
		eventChan: eventChan,
	}

	return h, nil
}

func (h *EventHandler) Handle(e *Message, ec chan error, args ...interface{}) {
	h.eventChan <- e
}
