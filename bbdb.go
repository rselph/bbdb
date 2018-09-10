package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultDBFile = "drive_stats.db"
)

var (
	dbFile  string
	wipe    bool
	driveDB *db
)

func main() {
	var err error

	flag.StringVar(&dbFile, "db", defaultDBFile, "Database file")
	flag.BoolVar(&wipe, "delete", false, "Delete old database before starting")
	flag.Parse()

	if wipe {
		_, err := os.Stat(dbFile)
		if err == nil {
			err = os.Remove(dbFile)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	driveDB, err = newDB(dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer driveDB.close()

	for _, dir := range flag.Args() {
		readOneDir(dir)
	}
}

func readOneDir(dir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, inErr error) (outErr error) {
		if info != nil &&
			!info.IsDir() &&
			strings.HasSuffix(strings.ToLower(info.Name()), ".csv") {
			err := readOneFile(path)
			if err != nil {
				log.Println(err)
			}
		}
		return
	})
}

func readOneFile(fname string) (err error) {
	log.Println(fname)
	s, err := openFile(fname)
	if err != nil {
		return
	}
	defer s.Close()

	return
}
