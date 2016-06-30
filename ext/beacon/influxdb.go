package beacon

import (
	"fmt"
	"strings"

	"github.com/ehazlett/interlock/config"
	influx "github.com/influxdata/influxdb/client/v2"
)

// queryDB convenience function to query the database
func queryDB(clnt influx.Client, db, cmd string) (res []influx.Result, err error) {
	q := influx.Query{
		Command:  cmd,
		Database: db,
	}
	if response, err := clnt.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}

	return res, nil
}

func NewInfluxDBClient(cfg *config.ExtensionConfig) (influx.Client, error) {
	c, err := influx.NewHTTPClient(
		influx.HTTPConfig{
			Addr:     cfg.StatsInfluxDBAddress,
			Username: cfg.StatsInfluxDBUser,
			Password: cfg.StatsInfluxDBPassword,
		},
	)
	if err != nil {
		return nil, err
	}

	// check db; create if needed
	dbName := cfg.StatsInfluxDBDatabase
	if _, err := queryDB(c, dbName, "SELECT * from cpu_usage limit 1"); err != nil {
		log().Debugf("error checking influx database: %q", err.Error())
		if strings.Index(err.Error(), "database not found") != -1 {
			// create database if needed
			log().Infof("creating stats influx database: %s", dbName)
			if _, err := queryDB(c, dbName, fmt.Sprintf("CREATE DATABASE %s", dbName)); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return c, nil
}
