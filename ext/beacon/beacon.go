package beacon

import (
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/ehazlett/interlock/config"
	"github.com/samalba/dockerclient"
)

const (
	pluginName = "beacon"
)

var (
	errChan chan (error)
)

type Beacon struct {
	cfg    *config.ExtensionConfig
	client *dockerclient.DockerClient
}

func log() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"ext": pluginName,
	})
}

type eventArgs struct {
	Image string
}

func NewBeacon(c *config.ExtensionConfig, client *dockerclient.DockerClient) (*Beacon, error) {
	// parse config base dir
	c.ConfigBasePath = filepath.Dir(c.ConfigPath)

	errChan = make(chan error)
	go func() {
		err := <-errChan
		log().Error(err)
	}()

	ext := &Beacon{
		cfg:    c,
		client: client,
	}

	return ext, nil
}

func (b *Beacon) Update() error {
	// update is not handled by beacon
	return nil
}

func (b *Beacon) Reload() error {
	// reload is not handled by beacon
	return nil
}

func (b *Beacon) HandleEvent(event *dockerclient.Event) error {
	switch event.Status {
	case "start":
		// get container info for event
		c, err := b.client.InspectContainer(event.ID)
		if err != nil {
			return err
		}

		image := c.Config.Image

		// TODO: match rules
		if !b.ruleMatch(c.Config) {
			return nil
		}

		log().Debugf("starting collection: id=%s image=%s", event.ID, image)
		if err := b.startStats(event.ID, image); err != nil {
			return err
		}
	case "kill", "die", "stop", "destroy":
		c, err := b.client.InspectContainer(event.ID)
		if err != nil {
			return err
		}

		image := c.Config.Image
		log().Debugf("resetting stats: id=%s image=%s", event.ID, image)
		if err := b.resetStats(event.ID, image); err != nil {
			return err
		}
	}

	return nil
}

func (b *Beacon) ruleMatch(cfg *dockerclient.ContainerConfig) bool {
	// TODO
	return true
}

func (b *Beacon) startStats(id string, image string) error {
	log().Debugf("gathering stats: id=%s image=%s interval=%d", id, image, b.cfg.StatInterval)
	args := eventArgs{
		Image: image,
	}
	go b.handleStats(id, b.sendContainerStats, errChan, args)

	return nil
}

func (b *Beacon) handleStats(id string, cb dockerclient.StatCallback, ec chan error, args ...interface{}) {
	go b.client.StartMonitorStats(id, cb, ec, args...)
}
