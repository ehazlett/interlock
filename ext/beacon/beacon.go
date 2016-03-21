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
		for err := range errChan {
			log().Error(err)
		}
	}()

	ext := &Beacon{
		cfg:    c,
		client: client,
	}

	return ext, nil
}

func (b *Beacon) Name() string {
	return pluginName
}

func (b *Beacon) HandleEvent(event *dockerclient.Event) error {
	switch event.Status {
	case "interlock-start":
		// scan all containers and start metrics
		containers, err := b.client.ListContainers(false, false, "")
		if err != nil {
			return err
		}

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
			return err
		}
	}

	return nil
}
