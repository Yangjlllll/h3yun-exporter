package collector

import (
	"fmt"
	"h3yun-scraper/api"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	serviceVesselMapCollectorName = "servicemap"
	serviceVesselMapLabelName     = []string{
		"slice", "vessel", "service_name", "service_type",
	}
)

type serviceMapCollector struct {
	serviceMapDesc *prometheus.Desc
	logger         log.Logger
}

func init() {
	registerCollector(serviceVesselMapCollectorName, defaultEnabled, NewServiceMapCollector)
}

func NewServiceMapCollector(logger log.Logger) (Collector, error) {
	smc := &serviceMapCollector{
		serviceMapDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, serviceVesselMapCollectorName, ""),
			"h3yun service service distribution map on vesel",
			serviceVesselMapLabelName, nil,
		),
		logger: logger,
	}
	return smc, nil
}

func (smc *serviceMapCollector) Update(ch chan<- prometheus.Metric) error {
	svMap, err := api.GetVesselServiceMap()
	if err != nil {
		return err
	}
	for _, serviceInfo := range svMap {
		ch <- prometheus.MustNewConstMetric(
			smc.serviceMapDesc, prometheus.GaugeValue,
			serviceInfo.Slice, fmt.Sprintf("%f", serviceInfo.Slice),
			serviceInfo.ShardKey, serviceInfo.Id, serviceInfo.ServiceName,
		)
	}
	return nil
}
