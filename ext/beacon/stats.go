package beacon

import (
	"math"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/samalba/dockerclient"
)

func (b *Beacon) sendContainerStats(id string, stats *dockerclient.Stats, ec chan error, args ...interface{}) {
	// report on interval
	timestamp := time.Now()

	rem := math.Mod(float64(timestamp.Second()), float64(b.cfg.StatInterval))
	if rem != 0 {
		return
	}

	log().Debugf("updating container stats: id=%s", id)

	image := ""
	if len(args) > 0 {
		arg := args[0]
		evtArgs := arg.(eventArgs)
		image = evtArgs.Image
	}

	if len(id) >= 12 {
		id = id[:12]
	}

	allContainers, err := b.client.ListContainers(true, false, "")
	if err != nil {
		log().Errorf("unable to list containers: %s", err)
		return
	}

	cInfo, err := b.client.InspectContainer(id)
	if err != nil {
		log().Errorf("unable to inspect container: %s", err)
		return
	}

	cName := cInfo.Name

	// strip /
	if cName[0] == '/' {
		cName = cName[1:]
	}

	counterTotalContainers.With(prometheus.Labels{
		"type": "totals",
	}).Set(float64(len(allContainers)))

	allImages, err := b.client.ListImages(true)
	if err != nil {
		log().Errorf("unable to list images: %s", err)
		return
	}

	counterTotalImages.With(prometheus.Labels{
		"type": "totals",
	}).Set(float64(len(allImages)))

	allVolumes, err := b.client.ListVolumes()
	if err != nil {
		log().Errorf("unable to list volumes: %s", err)
		return
	}

	networks, err := b.client.ListNetworks("")
	if err != nil {
		log().Errorf("unable to list networks: %s", err)
		return
	}
	counterTotalNetworks.With(prometheus.Labels{
		"type": "totals",
	}).Set(float64(len(networks)))

	counterTotalVolumes.With(prometheus.Labels{
		"type": "totals",
	}).Set(float64(len(allVolumes)))

	totalUsage := stats.CpuStats.CpuUsage.TotalUsage
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

	counterNetworkRxBytes.With(prometheus.Labels{
		"container": id,
		"image":     image,
		"name":      cName,
		"type":      "network",
	}).Set(float64(stats.NetworkStats.RxBytes))

	counterNetworkRxPackets.With(prometheus.Labels{
		"container": id,
		"image":     image,
		"name":      cName,
		"type":      "network",
	}).Set(float64(stats.NetworkStats.RxPackets))

	counterNetworkRxErrors.With(prometheus.Labels{
		"container": id,
		"image":     image,
		"name":      cName,
		"type":      "network",
	}).Set(float64(stats.NetworkStats.RxErrors))

	counterNetworkRxDropped.With(prometheus.Labels{
		"container": id,
		"image":     image,
		"name":      cName,
		"type":      "network",
	}).Set(float64(stats.NetworkStats.RxDropped))

	counterNetworkTxBytes.With(prometheus.Labels{
		"container": id,
		"image":     image,
		"name":      cName,
		"type":      "network",
	}).Set(float64(stats.NetworkStats.TxBytes))

	counterNetworkTxPackets.With(prometheus.Labels{
		"container": id,
		"image":     image,
		"name":      cName,
		"type":      "network",
	}).Set(float64(stats.NetworkStats.TxPackets))

	counterNetworkTxErrors.With(prometheus.Labels{
		"container": id,
		"image":     image,
		"name":      cName,
		"type":      "network",
	}).Set(float64(stats.NetworkStats.TxErrors))

	counterNetworkTxDropped.With(prometheus.Labels{
		"container": id,
		"image":     image,
		"name":      cName,
		"type":      "network",
	}).Set(float64(stats.NetworkStats.TxDropped))
}

func (b *Beacon) resetStats(id, image string) error {
	for _, c := range allCounters {
		c.Reset()
	}

	return nil
}
