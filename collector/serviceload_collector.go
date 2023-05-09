package collector

import (
	"h3yun-scraper/api"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	serviceloadCollectorName = "serviceload"
	serviceloadLabelNames    = []string{"resource", "service", "service_class", "ip", "vesselid"}
)

func init() {
	registerCollector(serviceloadCollectorName, defaultEnabled, NewServiceloadCollect)
}

type serviceloadCollector struct {
	cpuDesc     *prometheus.Desc
	memDesc     *prometheus.Desc
	threadsDesc *prometheus.Desc
	logger      log.Logger
}

type serviceloadLabels struct {
	resource, service, service_class, ip, vesselid string
}

type serviceloadStats struct {
	labels                            serviceloadLabels
	cpu_usage, mem_usage, threads_num api.AgentGauge
}

func NewServiceloadCollect(logger log.Logger) (Collector, error) {
	cpuDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, serviceloadCollectorName, "cpu_usage"),
		"h3yun serviceload cpu usage percent",
		serviceloadLabelNames, nil,
	)
	memDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, serviceloadCollectorName, "mem_usage"),
		"h3yun serviceload mem usage percent",
		serviceloadLabelNames, nil,
	)
	threadsDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, serviceloadCollectorName, "threads_num"),
		"h3yun service threads count",
		serviceloadLabelNames, nil,
	)
	slc := &serviceloadCollector{
		cpuDesc:     cpuDesc,
		memDesc:     memDesc,
		threadsDesc: threadsDesc,
		logger:      logger,
	}
	return slc, nil
}

func (slc *serviceloadCollector) GetStats() ([]serviceloadStats, error) {
	var val string
	sl := []serviceloadStats{}
	hosts, err := api.GetRegisteredServiceIPs()
	if err != nil {
		return nil, err
	}
	vessel_map, err := api.GetVesselServiceMap()
	if err != nil {
		return nil, err
	}
	for _, host := range hosts {
		service_info_list, err := api.GetAgentServiceload(host)
		if err != nil {
			return nil, err
		}
		for _, service_info := range service_info_list {
			_, ok := vessel_map[service_info.Id]
			if !ok {
				val = ""
			} else {
				val = vessel_map[service_info.Id].ShardKey
			}
			sl = append(sl, serviceloadStats{
				labels: serviceloadLabels{
					resource:      "service",
					service:       service_info.Id,
					service_class: service_info.ClassName,
					ip:            host,
					vesselid:      val,
				},
				cpu_usage:   service_info.Gauges["cpu_usage_percent"],
				mem_usage:   service_info.Gauges["threads_num"],
				threads_num: service_info.Gauges["threads_num"],
			})
		}
	}
	return sl, nil
}

func (slc *serviceloadCollector) Update(ch chan<- prometheus.Metric) error {
	stats, err := slc.GetStats()
	if err != nil {
		return err
	}
	seen := map[serviceloadLabels]bool{}
	for _, s := range stats {
		if seen[s.labels] {
			continue
		}
		ch <- prometheus.NewMetricWithTimestamp(s.cpu_usage.Time,
			prometheus.MustNewConstMetric(
				slc.cpuDesc, prometheus.GaugeValue,
				s.cpu_usage.Value, s.labels.resource, s.labels.service,
				s.labels.service_class,
				s.labels.ip, s.labels.vesselid,
			))
		ch <- prometheus.NewMetricWithTimestamp(s.mem_usage.Time,
			prometheus.MustNewConstMetric(
				slc.memDesc, prometheus.GaugeValue,
				s.mem_usage.Value, s.labels.resource, s.labels.service,
				s.labels.service_class,
				s.labels.ip, s.labels.vesselid,
			))
		ch <- prometheus.NewMetricWithTimestamp(s.threads_num.Time,
			prometheus.MustNewConstMetric(
				slc.threadsDesc, prometheus.GaugeValue,
				s.threads_num.Value, s.labels.resource, s.labels.service,
				s.labels.service_class,
				s.labels.ip, s.labels.vesselid,
			))
	}
	return nil
}
