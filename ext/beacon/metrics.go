package beacon

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	counterTotalContainers = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "beacon",
			Subsystem: "docker",
			Name:      "total_containers",
			Help:      "Total number of containers",
		},
		[]string{
			"type",
		},
	)
	counterTotalImages = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "beacon",
			Subsystem: "docker",
			Name:      "total_images",
			Help:      "Total number of images",
		},
		[]string{
			"type",
		},
	)
	counterTotalVolumes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "beacon",
			Subsystem: "docker",
			Name:      "total_volumes",
			Help:      "Total number of volumes",
		},
		[]string{
			"type",
		},
	)
	counterTotalNetworks = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "beacon",
			Subsystem: "docker",
			Name:      "total_networks",
			Help:      "Total number of networks",
		},
		[]string{
			"type",
		},
	)
	counterCpuTotalUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "beacon",
			Subsystem: "docker",
			Name:      "cpu_total_time_nanoseconds",
			Help:      "Total CPU time used in nanoseconds",
		},
		[]string{
			"container",
			"image",
			"name",
			"type",
		},
	)
	counterMemoryUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "beacon",
			Subsystem: "docker",
			Name:      "memory_usage_bytes",
			Help:      "Memory used in bytes",
		},
		[]string{
			"container",
			"image",
			"name",
			"type",
		},
	)
	counterMemoryMaxUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "beacon",
			Subsystem: "docker",
			Name:      "memory_max_usage_bytes",
			Help:      "Memory max used in bytes",
		},
		[]string{
			"container",
			"image",
			"name",
			"type",
		},
	)
	counterMemoryPercent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "beacon",
			Subsystem: "docker",
			Name:      "memory_usage_percent",
			Help:      "Percentage of memory used",
		},
		[]string{
			"container",
			"image",
			"name",
			"type",
		},
	)
	counterNetworkRxBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "beacon",
			Subsystem: "docker",
			Name:      "network_rx_bytes",
			Help:      "Network (rx) in bytes",
		},
		[]string{
			"container",
			"image",
			"name",
			"type",
		},
	)
	counterNetworkRxPackets = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "beacon",
			Subsystem: "docker",
			Name:      "network_rx_packets_total",
			Help:      "Network (rx) packet total",
		},
		[]string{
			"container",
			"image",
			"name",
			"type",
		},
	)
	counterNetworkRxErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "beacon",
			Subsystem: "docker",
			Name:      "network_rx_errors_total",
			Help:      "Network (rx) error total",
		},
		[]string{
			"container",
			"image",
			"name",
			"type",
		},
	)
	counterNetworkRxDropped = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "beacon",
			Subsystem: "docker",
			Name:      "network_rx_dropped_total",
			Help:      "Network (rx) dropped total",
		},
		[]string{
			"container",
			"image",
			"name",
			"type",
		},
	)
	counterNetworkTxBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "beacon",
			Subsystem: "docker",
			Name:      "network_tx_bytes",
			Help:      "Network (tx) in bytes",
		},
		[]string{
			"container",
			"image",
			"name",
			"type",
		},
	)
	counterNetworkTxPackets = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "beacon",
			Subsystem: "docker",
			Name:      "network_tx_packets_total",
			Help:      "Network (tx) packet total",
		},
		[]string{
			"container",
			"image",
			"name",
			"type",
		},
	)
	counterNetworkTxErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "beacon",
			Subsystem: "docker",
			Name:      "network_tx_errors_total",
			Help:      "Network (tx) error total",
		},
		[]string{
			"container",
			"image",
			"name",
			"type",
		},
	)
	counterNetworkTxDropped = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "beacon",
			Subsystem: "docker",
			Name:      "network_tx_dropped_total",
			Help:      "Network (tx) dropped total",
		},
		[]string{
			"container",
			"image",
			"name",
			"type",
		},
	)

	allCounters = []*prometheus.GaugeVec{
		counterTotalContainers,
		counterTotalImages,
		counterTotalVolumes,
		counterTotalNetworks,
		counterCpuTotalUsage,
		counterMemoryUsage,
		counterMemoryMaxUsage,
		counterMemoryPercent,
		counterNetworkRxBytes,
		counterNetworkRxPackets,
		counterNetworkRxErrors,
		counterNetworkRxDropped,
		counterNetworkTxBytes,
		counterNetworkTxPackets,
		counterNetworkTxErrors,
		counterNetworkTxDropped,
	}
)

func init() {
	// register the prometheus counters
	for _, c := range allCounters {
		prometheus.MustRegister(c)
	}
}
