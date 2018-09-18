package main

import (
	"encoding/csv"
	"errors"
	"os"
)

type smartFile struct {
	columns []string
	rows    [][]string
}

func readFile(fname string) (s *smartFile, err error) {
	r, err := os.Open(fname)
	if err != nil {
		return
	}
	defer r.Close()

	csvReader := csv.NewReader(r)
	var all [][]string
	all, err = csvReader.ReadAll()
	if err != nil {
		return
	}

	if len(all) < 2 {
		err = errors.New("Not enough rows")
	} else {
		s = &smartFile{columns: all[0], rows: all[1:]}
	}

	return
}

func (s *smartFile) rowCount() int {
	return len(s.rows)
}
