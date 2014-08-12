package main

import (
	"time"

	"github.com/citadel/citadel"
)

type (
	EventHandler struct {
		Manager *Manager
	}
)

func (l *EventHandler) Handle(e *citadel.Event) error {
	logger.Infof("event: date=%s type=%s image=%s container=%s", e.Time.Format(time.RubyDate), e.Type, e.Container.Image.Name, e.Container.ID[:12])
	switch e.Type {
	case "start":
		l.handleUpdate()
	case "kill":
		// add delay to make sure container is removed
		time.Sleep(250 * time.Millisecond)
		l.handleUpdate()
	}
	return nil
}

func (l *EventHandler) handleUpdate() error {
	if err := l.Manager.UpdateConfig(); err != nil {
		logger.Fatal(err)
	}
	if err := l.Manager.Reload(); err != nil {
		return err
	}
	return nil
}
