package main

import (
	"flag"
	"log"
	"os"
)

const (
	defaultDBFile = "drive_stats.db"
)

var (
	dbFile string
	wipe bool
)

func main() {
	var err error

	flag.StringVar(&dbFile, "db", defaultDBFile, "Database file")
	flag.BoolVar(&wipe, "delete", false, "Delete old database before starting")
	flag.Parse()

	if wipe {
		err = os.Remove(dbFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	_, err = newDB(dbFile)
	if err != nil {
		log.Fatal(err)
	}
}
