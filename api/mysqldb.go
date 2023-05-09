package api

import (
	"fmt"
	"h3yun-scraper/config"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	engineConfigDbName = "h_engineconfig"
	vesselConfigDbName = "h_vesselconfig"
)

type HAgentV2Model struct {
	Id   int    `gorm:"column:id"`
	Ip   string `gorm:"column:ip"`
	Port int    `gorm:"column:port"`
}

type VesselUnitstateData struct {
	VesselCode string  `gorm:"column:vesselcode"`
	Count      float64 `gorm:"column:count"`
}

type DbinstanceUnitstateData struct {
	DBInstanceId string  `gorm:"column:dbinstanceid"`
	VesselCode   string  `gorm:"column:vesselcode"`
	Count        float64 `gorm:"column:count"`
}

type SuiteKeyCountData struct {
	SuiteKey string
	Count    float64
}

func InitDb(dbconf config.DBConfig) (*gorm.DB, error) {
	dsn := dbconf.Username + ":" + dbconf.Password + "@tcp(" + dbconf.Host + ":" + dbconf.Port + ")/" + dbconf.Dbname
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func GetHagentv2Data() ([]HAgentV2Model, error) {
	var datas []HAgentV2Model
	db, err := InitDb(config.GlobalConfig.DBConfig)
	if err != nil {
		return nil, err
	}
	err = db.Table("h_agent_v2").Select("id, ip, port").Find(&datas).Error
	if err != nil {
		return nil, err
	}
	return datas, nil
}

func SelectVesselByUnitstate(unitstate int) ([]VesselUnitstateData, error) {
	var data []VesselUnitstateData
	db, err := InitDb(config.GlobalConfig.SharedServiceDBConfig)
	if err != nil {
		return nil, err
	}
	err = db.Table(engineConfigDbName).
		Select("vesselcode, COUNT(vesselcode) as count").
		Where("unitstate = ?", unitstate).
		Group("vesselcode").
		Find(&data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func SelectDbinstanceIdByUnitstate(unitstate int) ([]DbinstanceUnitstateData, error) {
	var data []DbinstanceUnitstateData
	db, err := InitDb(config.GlobalConfig.SharedServiceDBConfig)
	if err != nil {
		return nil, err
	}
	err = db.Table(engineConfigDbName).
		Select("dbinstanceid, COUNT(dbinstanceid), vesselcode").
		Where("unitstate = ?", unitstate).
		Find(&data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func GetVesselBySuiteKey(suiteKey string) ([]string, error) {
	data := []string{}
	db, err := InitDb(config.GlobalConfig.SharedServiceDBConfig)
	if err != nil {
		return nil, err
	}
	err = db.Table(vesselConfigDbName).
		Select("Code").
		Where("SuiteKey = ?", suiteKey).
		Find(&data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func GetAvailableDbCount() ([]SuiteKeyCountData, error) {
	data := []SuiteKeyCountData{}
	db, err := InitDb(config.GlobalConfig.SharedServiceDBConfig)
	if err != nil {
		return nil, err
	}
	db.Table(engineConfigDbName).
		Select("h_vesselconfig.SuiteKey, COUNT(h_engineconfig.code) as count").
		Joins("LEFT OUTER JOIN h_vesselconfig ON h_engineconfig.VesselCode = h_vesselconfig.Code").
		Where("h_vesselconfig.SuiteKey IS NOT NULL AND h_engineconfig.UnitState = ?", 2).
		Group("h_vesselconfig.SuiteKey").
		Find(&data)
	return data, err
}

type unitstateCache struct {
	result    []VesselUnitstateData
	err       error
	timestamp time.Time
}

func SelectVesselByUnitstateCache(ttl time.Duration, fn func(int) ([]VesselUnitstateData, error)) func(int) ([]VesselUnitstateData, error) {
	var m = map[int]unitstateCache{}
	return func(n int) ([]VesselUnitstateData, error) {
		if e, ok := m[n]; ok && time.Since(e.timestamp) < ttl {
			fmt.Println("Using cached result")
			return e.result, e.err
		}
		ret, err := fn(n)
		m[n] = unitstateCache{result: ret, err: err, timestamp: time.Now()}
		return ret, err
	}
}
