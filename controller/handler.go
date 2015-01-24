package main

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"
)

type (
	EventHandler struct {
		Manager *Manager
	}
)

func NewEventHandler(mgr *Manager) *EventHandler {
	return &EventHandler{
		Manager: mgr,
	}
}

func (l *EventHandler) Handle(e *dockerclient.Event, ec chan error, args ...interface{}) {
	log.Infof("event: date=%d type=%s image=%s container=%s", e.Time, e.Status, e.From, e.Id[:12])
	switch e.Status {
	case "start", "restart":
		l.handleUpdate(e)
	case "stop", "kill", "die":
		// add delay to make sure container is removed
		time.Sleep(250 * time.Millisecond)
		l.handleUpdate(e)
	}
}

func (l *EventHandler) handleUpdate(e *dockerclient.Event) error {
	if err := l.Manager.UpdateConfig(e); err != nil {
		return err
	}
	if err := l.Manager.Reload(); err != nil {
		return err
	}
	return nil
}
