package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/renevo/mcstats/pkg/db"
	"github.com/renevo/mcstats/pkg/influx"
	"github.com/renevo/mcstats/pkg/logging"
	"github.com/renevo/mcstats/pkg/mcstats"
)

func main() {
	statsDirectory := flag.String("dir", "./stats", "Minecraft stats directory")
	influxServer := flag.String("dbserver", "http://localhost:8086", "influx server to connect to")
	influxDB := flag.String("db", "minecraft", "database name")

	dataPath := flag.String("data-dir", "./data", "The location to store data")

	flag.Parse()

	db, err := db.Open(path.Join(*dataPath, "minecraft.db"))
	logging.Panic(err)
	defer db.Close()

	var lastStart int64
	logging.Panic(db.Get("minecraft", "start", &lastStart))
	logging.Debug("Last Start: %v", time.Unix(0, lastStart))
	logging.Panic(db.Put("minecraft", "start", time.Now().UnixNano()))

	idb, err := influx.CreateDatabase(*influxServer, *influxDB)
	logging.Panic(err)

	if _, err := os.Stat(*statsDirectory); os.IsNotExist(err) {
		logging.Panic(fmt.Errorf("directory %q does not exist", *statsDirectory))
	}

	statsFiles, err := mcstats.Load(*statsDirectory)
	logging.Panic(err)

	for _, stats := range statsFiles {
		if err := idb.Write(stats.ToLineFormat()); err != nil {
			logging.Error("%v", err)
			continue
		}
	}

	logging.Debug("Work Complete")
}
