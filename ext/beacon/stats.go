package beacon

import (
	"bufio"
	"encoding/json"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
)

type Stat struct {
	ID              string
	Name            string
	Image           string
	ContainerJSON   types.ContainerJSON
	NumContainers   int
	NumImages       int
	NumVolumes      int
	NumNetworks     int
	CPUTotalUsage   uint64
	MemUsagePercent float64
	Networks        map[string]types.NetworkStats
	Stats           *types.StatsJSON
}

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
	image = cInfo.Config.Image

	// strip /
	if cName[0] == '/' {
		cName = cName[1:]
	}

	imgOpts := types.ImageListOptions{
		All: true,
	}
	allImages, err := b.client.ImageList(context.Background(), imgOpts)
	if err != nil {
		log().Errorf("unable to list images: %s", err)
		return
	}

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

	totalUsage := stats.CPUStats.CPUUsage.TotalUsage
	memPercent := float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit) * 100.0
	numContainers := len(allContainers)
	numImages := len(allImages)
	numVolumes := len(allVolumes.Volumes)
	numNetworks := len(networks)

	s := Stat{
		ID:              id,
		Image:           image,
		Name:            cName,
		ContainerJSON:   cInfo,
		NumContainers:   numContainers,
		NumImages:       numImages,
		NumVolumes:      numVolumes,
		NumNetworks:     numNetworks,
		CPUTotalUsage:   totalUsage,
		MemUsagePercent: memPercent,
		Networks:        stats.Networks,
		Stats:           stats,
	}

	log().Debugf("stats backend: type=%s", b.cfg.StatsBackendType)

	switch b.cfg.StatsBackendType {
	case "prometheus":
		log().Debug("updating rometheus")
		b.sendPrometheus(s)
	case "influxdb":
		log().Debugf("sending stats to influxdb: url=%s", b.cfg.StatsInfluxDBAddress)
		// TODO
		if err := b.sendInfluxDB(s); err != nil {
			log().Errorf("error sending stats to influxdb: %s", err)
			return
		}
	default:
		log().Errorf("unsupported stat backend: %s", b.cfg.StatsBackendType)
		return
	}
}

func (b *Beacon) sendPrometheus(stat Stat) {
	counterTotalContainers.With(prometheus.Labels{
		"type": "totals",
	}).Set(float64(stat.NumContainers))

	counterTotalImages.With(prometheus.Labels{
		"type": "totals",
	}).Set(float64(stat.NumImages))

	counterTotalNetworks.With(prometheus.Labels{
		"type": "totals",
	}).Set(float64(stat.NumNetworks))

	counterTotalVolumes.With(prometheus.Labels{
		"type": "totals",
	}).Set(float64(stat.NumVolumes))

	counterCpuTotalUsage.With(prometheus.Labels{
		"container": stat.ID,
		"image":     stat.Image,
		"name":      stat.Name,
		"type":      "cpu",
	}).Set(float64(stat.CPUTotalUsage))

	counterMemoryUsage.With(prometheus.Labels{
		"container": stat.ID,
		"image":     stat.Image,
		"name":      stat.Name,
		"type":      "memory",
	}).Set(float64(stat.Stats.MemoryStats.Usage))

	counterMemoryMaxUsage.With(prometheus.Labels{
		"container": stat.ID,
		"image":     stat.Image,
		"name":      stat.Name,
		"type":      "memory",
	}).Set(float64(stat.Stats.MemoryStats.MaxUsage))

	counterMemoryPercent.With(prometheus.Labels{
		"container": stat.ID,
		"image":     stat.Image,
		"name":      stat.Name,
		"type":      "memory",
	}).Set(float64(stat.MemUsagePercent))

	for netName, net := range stat.Networks {
		counterNetworkRxBytes.With(prometheus.Labels{
			"container": stat.ID,
			"image":     stat.Image,
			"name":      stat.Name,
			"network":   netName,
			"type":      "network",
		}).Set(float64(net.RxBytes))

		counterNetworkRxPackets.With(prometheus.Labels{
			"container": stat.ID,
			"image":     stat.Image,
			"name":      stat.Name,
			"network":   netName,
			"type":      "network",
		}).Set(float64(net.RxPackets))

		counterNetworkRxErrors.With(prometheus.Labels{
			"container": stat.ID,
			"image":     stat.Image,
			"name":      stat.Name,
			"network":   netName,
			"type":      "network",
		}).Set(float64(net.RxErrors))

		counterNetworkRxDropped.With(prometheus.Labels{
			"container": stat.ID,
			"image":     stat.Image,
			"name":      stat.Name,
			"network":   netName,
			"type":      "network",
		}).Set(float64(net.RxDropped))

		counterNetworkTxBytes.With(prometheus.Labels{
			"container": stat.ID,
			"image":     stat.Image,
			"name":      stat.Name,
			"network":   netName,
			"type":      "network",
		}).Set(float64(net.TxBytes))

		counterNetworkTxPackets.With(prometheus.Labels{
			"container": stat.ID,
			"image":     stat.Image,
			"name":      stat.Name,
			"network":   netName,
			"type":      "network",
		}).Set(float64(net.TxPackets))

		counterNetworkTxErrors.With(prometheus.Labels{
			"container": stat.ID,
			"image":     stat.Image,
			"name":      stat.Name,
			"network":   netName,
			"type":      "network",
		}).Set(float64(net.TxErrors))

		counterNetworkTxDropped.With(prometheus.Labels{
			"container": stat.ID,
			"image":     stat.Image,
			"name":      stat.Name,
			"network":   netName,
			"type":      "network",
		}).Set(float64(net.TxDropped))
	}
}

func (b *Beacon) sendInfluxDB(stat Stat) error {
	c, err := NewInfluxDBClient(b.cfg)
	if err != nil {
		return err
	}

	bp, err := influx.NewBatchPoints(influx.BatchPointsConfig{
		Database:  b.cfg.StatsInfluxDBDatabase,
		Precision: b.cfg.StatsInfluxDBPrecision,
	})
	if err != nil {
		return err
	}

	timestamp := time.Now()

	// cpu total
	pt, err := influx.NewPoint("containers", map[string]string{
		"type": "totals",
	}, map[string]interface{}{
		"total_containers": stat.NumContainers,
	}, timestamp)
	if err != nil {
		return err
	}
	bp.AddPoint(pt)

	if err := c.Write(bp); err != nil {
		return err
	}

	// images total
	pt, err = influx.NewPoint("images", map[string]string{
		"type": "totals",
	}, map[string]interface{}{
		"total_images": stat.NumImages,
	}, timestamp)
	if err != nil {
		return err
	}
	bp.AddPoint(pt)

	// networks total
	pt, err = influx.NewPoint("networks", map[string]string{
		"type": "totals",
	}, map[string]interface{}{
		"total_networks": stat.NumNetworks,
	}, timestamp)
	if err != nil {
		return err
	}
	bp.AddPoint(pt)

	// volumes total
	pt, err = influx.NewPoint("volumes", map[string]string{
		"type": "totals",
	}, map[string]interface{}{
		"total_volumes": stat.NumVolumes,
	}, timestamp)
	if err != nil {
		return err
	}
	bp.AddPoint(pt)

	// cpu total
	pt, err = influx.NewPoint("cpu_usage", map[string]string{
		"cpu":       "cpu-total",
		"resource":  "container",
		"container": stat.ID,
		"image":     stat.Image,
		"name":      stat.Name,
	}, map[string]interface{}{
		"value": stat.CPUTotalUsage,
	}, timestamp)
	if err != nil {
		return err
	}
	bp.AddPoint(pt)

	// mem usage
	pt, err = influx.NewPoint("mem_usage", map[string]string{
		"memory":    "memory-usage",
		"resource":  "container",
		"container": stat.ID,
		"image":     stat.Image,
		"name":      stat.Name,
	}, map[string]interface{}{
		"value": stat.Stats.MemoryStats.Usage,
	}, timestamp)
	if err != nil {
		return err
	}
	bp.AddPoint(pt)

	// mem max usage
	pt, err = influx.NewPoint("mem_max_usage", map[string]string{
		"memory":    "memory-max-usage",
		"resource":  "container",
		"container": stat.ID,
		"image":     stat.Image,
		"name":      stat.Name,
	}, map[string]interface{}{
		"value": stat.Stats.MemoryStats.MaxUsage,
	}, timestamp)
	if err != nil {
		return err
	}
	bp.AddPoint(pt)

	// mem usage percent
	pt, err = influx.NewPoint("mem_usage_percent", map[string]string{
		"memory":    "memory-usage-percent",
		"resource":  "container",
		"container": stat.ID,
		"image":     stat.Image,
		"name":      stat.Name,
	}, map[string]interface{}{
		"value": stat.MemUsagePercent,
	}, timestamp)
	if err != nil {
		return err
	}
	bp.AddPoint(pt)

	// networks
	for netName, net := range stat.Networks {
		pt, err = influx.NewPoint("net_rx_bytes", map[string]string{
			"resource":  "network",
			"container": stat.ID,
			"network":   netName,
			"image":     stat.Image,
			"name":      stat.Name,
		}, map[string]interface{}{
			"value": net.RxBytes,
		}, timestamp)
		if err != nil {
			return err
		}
		bp.AddPoint(pt)

		pt, err = influx.NewPoint("net_rx_packets", map[string]string{
			"resource":  "network",
			"container": stat.ID,
			"network":   netName,
			"image":     stat.Image,
			"name":      stat.Name,
		}, map[string]interface{}{
			"value": net.RxPackets,
		}, timestamp)
		if err != nil {
			return err
		}
		bp.AddPoint(pt)

		pt, err = influx.NewPoint("net_rx_errors", map[string]string{
			"resource":  "network",
			"container": stat.ID,
			"network":   netName,
			"image":     stat.Image,
			"name":      stat.Name,
		}, map[string]interface{}{
			"value": net.RxErrors,
		}, timestamp)
		if err != nil {
			return err
		}
		bp.AddPoint(pt)

		pt, err = influx.NewPoint("net_rx_dropped", map[string]string{
			"resource":  "network",
			"container": stat.ID,
			"network":   netName,
			"image":     stat.Image,
			"name":      stat.Name,
		}, map[string]interface{}{
			"value": net.RxDropped,
		}, timestamp)
		if err != nil {
			return err
		}
		bp.AddPoint(pt)

		pt, err = influx.NewPoint("net_tx_bytes", map[string]string{
			"resource":  "network",
			"container": stat.ID,
			"network":   netName,
			"image":     stat.Image,
			"name":      stat.Name,
		}, map[string]interface{}{
			"value": net.TxBytes,
		}, timestamp)
		if err != nil {
			return err
		}
		bp.AddPoint(pt)

		pt, err = influx.NewPoint("net_tx_packets", map[string]string{
			"resource":  "network",
			"container": stat.ID,
			"network":   netName,
			"image":     stat.Image,
			"name":      stat.Name,
		}, map[string]interface{}{
			"value": net.TxPackets,
		}, timestamp)
		if err != nil {
			return err
		}
		bp.AddPoint(pt)

		pt, err = influx.NewPoint("net_tx_errors", map[string]string{
			"resource":  "network",
			"container": stat.ID,
			"network":   netName,
			"image":     stat.Image,
			"name":      stat.Name,
		}, map[string]interface{}{
			"value": net.TxErrors,
		}, timestamp)
		if err != nil {
			return err
		}
		bp.AddPoint(pt)

		pt, err = influx.NewPoint("net_tx_dropped", map[string]string{
			"resource":  "network",
			"container": stat.ID,
			"network":   netName,
			"image":     stat.Image,
			"name":      stat.Name,
		}, map[string]interface{}{
			"value": net.TxDropped,
		}, timestamp)
		if err != nil {
			return err
		}
		bp.AddPoint(pt)
	}

	if err := c.Write(bp); err != nil {
		return err
	}

	return nil
}

func (b *Beacon) collectStats() {
	wg := &sync.WaitGroup{}
	for id, _ := range b.monitored {
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
			s := bufio.NewScanner(r.Body)
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
