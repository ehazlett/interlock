package manager

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/ehazlett/interlock/plugins"
	"github.com/samalba/dockerclient"
)

// EventHandler dispatches events to the manager
type (
	EventHandler struct {
		Manager *Manager
	}
)

// NewEventHandler registers a Manager on a new EventHandler
func NewEventHandler(mgr *Manager) *EventHandler {
	return &EventHandler{
		Manager: mgr,
	}
}

// Handle sends events on to the Manager
func (l *EventHandler) Handle(e *dockerclient.Event, ec chan error, args ...interface{}) {
	plugins.Log("interlock", log.DebugLevel,
		fmt.Sprintf("event: date=%d type=%s image=%s container=%s", e.Time, e.Status, e.From, e.Id))

	go plugins.DispatchEvent(l.Manager.Config, l.Manager.Client, e, ec)
}
