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
<<<<<<< HEAD
		for err := range errChan {
			log().Error(err)
		}
=======
		err := <-errChan
		log().Error(err)
>>>>>>> af48d45... beacon extension
	}()

	ext := &Beacon{
		cfg:    c,
		client: client,
	}

	return ext, nil
}

<<<<<<< HEAD
func (b *Beacon) Name() string {
	return pluginName
=======
func (b *Beacon) Update() error {
	// update is not handled by beacon
	return nil
}

func (b *Beacon) Reload() error {
	// reload is not handled by beacon
	return nil
>>>>>>> af48d45... beacon extension
}

func (b *Beacon) HandleEvent(event *dockerclient.Event) error {
	switch event.Status {
<<<<<<< HEAD
	case "interlock-start":
		// scan all containers and start metrics
		containers, err := b.client.ListContainers(false, false, "")
=======
	case "start":
		// get container info for event
		c, err := b.client.InspectContainer(event.ID)
>>>>>>> af48d45... beacon extension
		if err != nil {
			return err
		}

<<<<<<< HEAD
		for _, c := range containers {
			if err := b.startStats(c.Id); err != nil {
				log().Warnf("unable to start stats for containers: id=%s err=%s", c.Id, err)
				continue
			}
		}
	case "start":
		log().Debugf("checking container for collection: id=%s", event.ID)
		if err := b.startStats(event.ID); err != nil {
			return err
		}
	case "kill", "die", "stop", "destroy":
		log().Debugf("resetting stats: id=%s", event.ID)
		if err := b.resetStats(event.ID); err != nil {
=======
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
>>>>>>> af48d45... beacon extension
			return err
		}
	}

	return nil
}
<<<<<<< HEAD
=======

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
>>>>>>> af48d45... beacon extension
