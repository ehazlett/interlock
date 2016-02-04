package events

import (
	log "github.com/Sirupsen/logrus"
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
	log.Debugf("raw event: id=%s type=%s", e.ID, e.Status)
	h.eventChan <- e
}
