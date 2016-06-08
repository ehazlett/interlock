package beacon

import (
	"bufio"
	"encoding/json"
	"sync"

	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
)

func (b *Beacon) sendContainerStats(id string, stats *types.StatsJSON, ec chan error, args ...interface{}) {
	log().Debugf("updating container stats: id=%s", id)

	image := ""
	if len(id) >= 12 {
		id = id[:12]
	}

	opts := types.ContainerListOptions{
		All:  true,
		Size: false,
	}
	allContainers, err := b.client.ContainerList(context.Background(), opts)
	if err != nil {
		log().Errorf("unable to list containers: %s", err)
		return
	}

	cInfo, err := b.client.ContainerInspect(context.Background(), id)
	if err != nil {
		log().Errorf("unable to inspect container: %s", err)
		return
	}

	cName := cInfo.Name
	image = cInfo.Image

	// strip /
	if cName[0] == '/' {
		cName = cName[1:]
	}

	counterTotalContainers.With(prometheus.Labels{
		"type": "totals",
	}).Set(float64(len(allContainers)))

	imgOpts := types.ImageListOptions{
		All: true,
	}
	allImages, err := b.client.ImageList(context.Background(), imgOpts)
	if err != nil {
		log().Errorf("unable to list images: %s", err)
		return
	}

	counterTotalImages.With(prometheus.Labels{
		"type": "totals",
	}).Set(float64(len(allImages)))

	allVolumes, err := b.client.VolumeList(context.Background(), filters.Args{})
	if err != nil {
		log().Errorf("unable to list volumes: %s", err)
		return
	}

	networks, err := b.client.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		log().Errorf("unable to list networks: %s", err)
		return
	}
	counterTotalNetworks.With(prometheus.Labels{
		"type": "totals",
	}).Set(float64(len(networks)))

	counterTotalVolumes.With(prometheus.Labels{
		"type": "totals",
	}).Set(float64(len(allVolumes.Volumes)))

	totalUsage := stats.CPUStats.CPUUsage.TotalUsage
	memPercent := float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit) * 100.0

	counterCpuTotalUsage.With(prometheus.Labels{
		"container": id,
		"image":     image,
		"name":      cName,
		"type":      "cpu",
	}).Set(float64(totalUsage))

	counterMemoryUsage.With(prometheus.Labels{
		"container": id,
		"image":     image,
		"name":      cName,
		"type":      "memory",
	}).Set(float64(stats.MemoryStats.Usage))

	counterMemoryMaxUsage.With(prometheus.Labels{
		"container": id,
		"image":     image,
		"name":      cName,
		"type":      "memory",
	}).Set(float64(stats.MemoryStats.MaxUsage))

	counterMemoryPercent.With(prometheus.Labels{
		"container": id,
		"image":     image,
		"name":      cName,
		"type":      "memory",
	}).Set(float64(memPercent))

	for netName, net := range stats.Networks {
		counterNetworkRxBytes.With(prometheus.Labels{
			"container": id,
			"image":     image,
			"name":      cName,
			"network":   netName,
			"type":      "network",
		}).Set(float64(net.RxBytes))

		counterNetworkRxPackets.With(prometheus.Labels{
			"container": id,
			"image":     image,
			"name":      cName,
			"network":   netName,
			"type":      "network",
		}).Set(float64(net.RxPackets))

		counterNetworkRxErrors.With(prometheus.Labels{
			"container": id,
			"image":     image,
			"name":      cName,
			"network":   netName,
			"type":      "network",
		}).Set(float64(net.RxErrors))

		counterNetworkRxDropped.With(prometheus.Labels{
			"container": id,
			"image":     image,
			"name":      cName,
			"network":   netName,
			"type":      "network",
		}).Set(float64(net.RxDropped))

		counterNetworkTxBytes.With(prometheus.Labels{
			"container": id,
			"image":     image,
			"name":      cName,
			"network":   netName,
			"type":      "network",
		}).Set(float64(net.TxBytes))

		counterNetworkTxPackets.With(prometheus.Labels{
			"container": id,
			"image":     image,
			"name":      cName,
			"network":   netName,
			"type":      "network",
		}).Set(float64(net.TxPackets))

		counterNetworkTxErrors.With(prometheus.Labels{
			"container": id,
			"image":     image,
			"name":      cName,
			"network":   netName,
			"type":      "network",
		}).Set(float64(net.TxErrors))

		counterNetworkTxDropped.With(prometheus.Labels{
			"container": id,
			"image":     image,
			"name":      cName,
			"network":   netName,
			"type":      "network",
		}).Set(float64(net.TxDropped))
	}
}

func (b *Beacon) collectStats() {
	wg := &sync.WaitGroup{}
	for id, _ := range b.monitored {
		log().Debugf("monitored: %v", b.monitored)
		log().Debugf("id: %s", id)
		wg.Add(1)
		go func(cID string) {
			defer wg.Done()

			c, err := b.client.ContainerInspect(context.Background(), cID)
			if err != nil {
				errChan <- err
				return
			}

			image := c.Config.Image

			// match rules
			if !b.ruleMatch(c.Config) {
				log().Debugf("unable to find rule matching container %s (%s); not monitoring", c.ID, image)
				return
			}

			log().Debugf("checking container stats: id=%s", cID)
			r, err := b.client.ContainerStats(context.Background(), cID, false)
			if err != nil {
				log().Errorf("unable to get container stats: %s", err)
				return
			}

			var stats *types.StatsJSON
			s := bufio.NewScanner(r)
			for s.Scan() {
				if err := json.Unmarshal([]byte(s.Text()), &stats); err != nil {
					log().Errorf("unable to unmarshal stats: %s", err)
					return
				}

				b.sendContainerStats(cID, stats, errChan)
			}

		}(id)
	}

	wg.Wait()
}

func (b *Beacon) resetStats(id string) error {
	for _, c := range allCounters {
		c.Reset()
	}

	return nil
}
