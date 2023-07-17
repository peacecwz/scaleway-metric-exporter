package collectors

import (
	"context"
	"github.com/peacecwz/scaleway-metric-exporter/config"
	"github.com/peacecwz/scaleway-metric-exporter/pkg/scaleway"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	log "github.com/sirupsen/logrus"
	"sort"
	"sync"
	"time"
)

type RDBCollector struct {
	cfg    *config.Config
	client *scaleway.ScalewayClient

	uptimeMetrics          *prometheus.Desc
	cpuMetrics             *prometheus.Desc
	memoryMetrics          *prometheus.Desc
	connectionCountMetrics *prometheus.Desc
	diskUsageMetrics       *prometheus.Desc
}

func NewRDBCollector(cfg *config.Config, scwClient *scaleway.ScalewayClient) *RDBCollector {
	log.Info("Creating new RDB collector")
	labels := []string{"id", "name", "region", "engine", "type"}

	labelsNode := []string{"id", "name", "node"}

	return &RDBCollector{
		cfg:    cfg,
		client: scwClient,

		uptimeMetrics: prometheus.NewDesc(
			"scaleway_rdb_uptime",
			"If 1 the database is up and running, 0.5 in auto healing, 0 otherwise",
			labels, nil,
		),
		cpuMetrics: prometheus.NewDesc(
			"scaleway_rdb_cpu_usage_percent",
			"RDB CPU percentage usage",
			labelsNode, nil,
		),
		memoryMetrics: prometheus.NewDesc(
			"scaleway_rdb_memory_usage_percent",
			"RDB memory percentage usage",
			labelsNode, nil,
		),
		connectionCountMetrics: prometheus.NewDesc(
			"scaleway_rdb_total_connections",
			"RDB connection count",
			labelsNode, nil,
		),
		diskUsageMetrics: prometheus.NewDesc(
			"scaleway_database_disk_usage_percent",
			"RDB disk percentage usage",
			labelsNode, nil,
		),
	}
}

func (c *RDBCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.uptimeMetrics
	ch <- c.cpuMetrics
	ch <- c.memoryMetrics
	ch <- c.connectionCountMetrics
	ch <- c.diskUsageMetrics
}

func (c *RDBCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.cfg.TimeoutInMs)*time.Millisecond)
	defer cancel()

	var wg sync.WaitGroup
	defer wg.Wait()

	for _, region := range c.client.Regions {
		response, err := c.client.ListRDBInstance(ctx, &rdb.ListInstancesRequest{Region: region}, scw.WithAllPages())
		if err != nil {
			log.Errorf("failed to fetch the list of databases: %v", err)
			return
		}

		log.Infof("found %d database instances", len(response.Instances))

		for _, instance := range response.Instances {
			wg.Add(1)

			log.Infof("Fetching metrics for database instance : %s", instance.Name)

			go c.FetchMetricsForInstance(&wg, ch, instance)
		}
	}
}

func (c *RDBCollector) FetchMetricsForInstance(parentWg *sync.WaitGroup, ch chan<- prometheus.Metric, instance *rdb.Instance) {
	defer parentWg.Done()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.cfg.TimeoutInMs)*time.Millisecond)
	defer cancel()

	labels := []string{
		instance.ID,
		instance.Name,
		instance.Region.String(),
		instance.Engine,
		instance.NodeType,
	}

	var active float64

	switch instance.Status {
	case rdb.InstanceStatusReady:
		active = 1.0
	case rdb.InstanceStatusBackuping:
		active = 1.0
	case rdb.InstanceStatusAutohealing:
		active = 0.5
	case rdb.InstanceStatusProvisioning:
		active = 0.5
	case rdb.InstanceStatusConfiguring:
		active = 0.5
	case rdb.InstanceStatusDeleting:
		active = 0.5
	case rdb.InstanceStatusSnapshotting:
		active = 0.5
	case rdb.InstanceStatusRestarting:
		active = 0.5
	case rdb.InstanceStatusUnknown:
		active = 0.0
	case rdb.InstanceStatusError:
		active = 0.0
	case rdb.InstanceStatusLocked:
		active = 0.0
	case rdb.InstanceStatusInitializing:
		active = 0.0
	case rdb.InstanceStatusDiskFull:
		active = 0.0
	default:
		active = 0.0
	}

	ch <- prometheus.MustNewConstMetric(
		c.uptimeMetrics,
		prometheus.GaugeValue,
		active,
		labels...,
	)

	metricResponse, err := c.client.GetInstanceMetrics(ctx, &rdb.GetInstanceMetricsRequest{Region: instance.Region, InstanceID: instance.ID})

	if err != nil {
		log.Warnf("failed to fetch metrics for database instance %s: %v", instance.Name, err)

		return
	}

	for _, metric := range metricResponse.Timeseries {
		labelsNode := []string{
			instance.ID,
			instance.Name,
			metric.Metadata["node"],
		}

		var series *prometheus.Desc

		switch metric.Name {
		case "cpu_usage_percent":
			series = c.cpuMetrics
		case "mem_usage_percent":
			series = c.memoryMetrics
		case "total_connections":
			series = c.connectionCountMetrics
		case "disk_usage_percent":
			series = c.diskUsageMetrics
		default:
			log.Warnf("unmapped scaleway metric: %s", metric.Name)
			continue
		}

		if len(metric.Points) == 0 {
			log.Warnf("no data were returned for the metric: %s", metric.Name)
			continue
		}

		sort.Slice(metric.Points, func(i, j int) bool {
			return metric.Points[i].Timestamp.Before(metric.Points[j].Timestamp)
		})

		value := float64(metric.Points[len(metric.Points)-1].Value)

		ch <- prometheus.MustNewConstMetric(series, prometheus.GaugeValue, value, labelsNode...)
	}
}
