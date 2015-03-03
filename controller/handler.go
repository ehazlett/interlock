package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/ehazlett/interlock/plugins"
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

	go plugins.DispatchEvent(l.Manager.Config, l.Manager.Client, e, ec)
	//switch e.Status {
	//case "start", "restart":
	//	l.handleUpdate(e)
	//case "stop", "kill", "die":
	//	// add delay to make sure container is removed
	//	time.Sleep(250 * time.Millisecond)
	//	l.handleUpdate(e)
	//}
}
