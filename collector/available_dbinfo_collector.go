package collector

import (
	"fmt"
	"h3yun-scraper/api"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	unitstateCode                      = 1
	availableDbCode                    = 2
	availableDbInfoCollectorName       = "availabledb"
	availableDbUnitstateCountLabelName = []string{"resource", "vessel", "status"}
	availableDbCountLabelName          = []string{}
	availableDbStatusCountLabelName    = []string{}
)

type availableDbInfoCollector struct {
	availableDbUnitstateCountDesc *prometheus.Desc
	availableDbCountDesc          *prometheus.Desc
	availableDbStatusCountDesc    *prometheus.Desc
	logger                        log.Logger
}

type availableDbCount struct {
	vesselType string
	suiteKey   string
	Count      float64
}

func init() {
	registerCollector(availableDbInfoCollectorName, defaultEnabled, NewAvailableDbInfoCollector)
}

func NewAvailableDbInfoCollector(logger log.Logger) (Collector, error) {
	unitstateCountDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, availableDbInfoCollectorName, "unitstate_count"),
		"h3yun vessel unitstate count",
		availableDbUnitstateCountLabelName, nil,
	)
	statusCountDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, availableDbInfoCollectorName, "status_count"),
		"h3yun dbinstanceid status count",
		availableDbStatusCountLabelName, nil,
	)
	availableDbCountDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, availableDbInfoCollectorName, "count"),
		"h3yun available db info",
		availableDbCountLabelName, nil,
	)
	adi := &availableDbInfoCollector{
		availableDbUnitstateCountDesc: unitstateCountDesc,
		availableDbCountDesc:          availableDbCountDesc,
		availableDbStatusCountDesc:    statusCountDesc,
		logger:                        logger,
	}
	return adi, nil
}

func (adi *availableDbInfoCollector) Update(ch chan<- prometheus.Metric) error {
	// fn := api.SelectVesselByUnitstateCache(30*time.Minute, api.SelectVesselByUnitstate)
	fn := api.FuncCache(30*time.Minute, api.SelectVesselByUnitstate).(func(int) ([]api.VesselUnitstateData, error))
	unitstateDbInfoList, err := fn(unitstateCode)
	if err != nil {
		return err
	}
	for _, unitstateDbInfo := range unitstateDbInfoList {
		ch <- prometheus.MustNewConstMetric(
			adi.availableDbUnitstateCountDesc, prometheus.GaugeValue,
			unitstateDbInfo.Count, "vessel", unitstateDbInfo.VesselCode, fmt.Sprintf("%d", unitstateCode),
		)
	}
	fn2 := api.FuncCache(30*time.Minute, api.SelectDbinstanceIdByUnitstate).(func(int) ([]api.DbinstanceUnitstateData, error))
	availableDBStateInfoList, err := fn2(availableDbCode)
	if err != nil {
		return err
	}
	for _, availableDBStateInfo := range availableDBStateInfoList {
		ch <- prometheus.MustNewConstMetric(
			adi.availableDbStatusCountDesc, prometheus.GaugeValue,
			availableDBStateInfo.Count, "rds", availableDBStateInfo.DBInstanceId,
			availableDBStateInfo.VesselCode, fmt.Sprintf("%d", availableDbCode),
		)
	}
	fn3 := api.FuncCache(15*time.Minute, getAvailableDbCountByState).(func() ([]availableDbCount, error))
	availableDBInfoList, err := fn3()
	if err != nil {
		return err
	}
	for _, availableDBInfo := range availableDBInfoList {
		ch <- prometheus.MustNewConstMetric(
			adi.availableDbCountDesc, prometheus.GaugeValue,
			availableDBInfo.Count, "available_db", availableDBInfo.vesselType,
			availableDBInfo.suiteKey,
		)
	}
	return nil
}

func getAvailableDbCountByState() ([]availableDbCount, error) {
	data := []availableDbCount{}
	availableDbCountInfoList, err := api.GetAvailableDbCount()
	if err != nil {
		return nil, err
	}
	for _, availableDbCountInfo := range availableDbCountInfoList {
		countInfo := availableDbCount{}
		if availableDbCountInfo.SuiteKey != "None" {
			suiteKey := availableDbCountInfo.SuiteKey
			vesselList, err := api.GetVesselBySuiteKey(suiteKey)
			if err != nil {
				return nil, err
			}
			countInfo.suiteKey = suiteKey + "_" + strings.Join(vesselList, "_")
			countInfo.vesselType = "saas_vessel"
			countInfo.Count = availableDbCountInfo.Count
		} else {
			countInfo.vesselType = "common_suitekey"
			countInfo.suiteKey = "common_suitekey"
			countInfo.Count = availableDbCountInfo.Count
		}
		data = append(data, countInfo)
	}
	return data, nil
}
