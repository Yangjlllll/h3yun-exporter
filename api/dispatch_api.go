package api

import (
	"encoding/json"
	"h3yun-scraper/config"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Service struct {
	Id      string `json:"Id"`
	Address string `json:"Address"`
}

type ServiceInfo struct {
	Id        string                `json:"id"`
	ClassName string                `json:"class_name"`
	Gauges    map[string]AgentGauge `json:"gauges"`
}

type AgentGauge struct {
	Value float64   `json:"value"`
	Time  time.Time `json:"time"`
}

type RegisterServiceInfo struct {
	Id          string  `json:"Id"`
	ServiceName string  `json:"ServiceName"`
	ShardKey    string  `json:"ShardKey"`
	Slice       float64 `json:"Slice"`
}

func GetRegisteredServiceIPs() ([]string, error) {
	url := config.GlobalConfig.DispatchURL
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rep map[string]map[string][]Service
	err = json.NewDecoder(resp.Body).Decode(&rep)
	if err != nil {
		return nil, err
	}

	var ipInfo []string
	for _, service := range rep["ReturnValue"]["ReturnValue"] {
		// 排除部分在k8s上运行且注册到Dispatch上的服务，这部分服务的指标不采用scraper方式
		r, err := regexp.Compile(".*-.*")
		if err != nil {
			return nil, err
		}
		if match := r.MatchString(service.Id); match {
			continue
		}
		ip := strings.Split(service.Address, ":")[0]
		if !contains(ipInfo, ip) {
			ipInfo = append(ipInfo, ip)
		}
	}
	return ipInfo, nil
}

func contains(strs []string, s string) bool {
	for _, str := range strs {
		if str == s {
			return true
		}
	}
	return false
}

func GetAgentServiceload(ip string) ([]ServiceInfo, error) {
	url := "http://" + ip + config.GlobalConfig.AgentRoute
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var rep map[string][]ServiceInfo
	err = json.NewDecoder(resp.Body).Decode(&rep)
	if err != nil {
		return nil, err
	}
	return rep["return"], nil
}

func GetVesselServiceMap() (map[string]RegisterServiceInfo, error) {
	url := config.GlobalConfig.RegisterInfoCacheUrl
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var service_info_list []RegisterServiceInfo
	vessel_map := make(map[string]RegisterServiceInfo)
	err = json.NewDecoder(resp.Body).Decode(&service_info_list)
	if err != nil {
		return nil, err
	}
	for _, service_info := range service_info_list {
		vessel_map[service_info.Id] = service_info
	}
	return vessel_map, nil
}
