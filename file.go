package main

import (
	"encoding/csv"
	"errors"
	"io"
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

	return readReader(r)
}

func readReader(r io.Reader) (s *smartFile, err error) {
	csvReader := csv.NewReader(r)
	var all [][]string
	all, err = csvReader.ReadAll()
	if err != nil {
		return
	}

	if len(all) < 2 {
		err = errors.New("not enough rows")
	} else {
		s = &smartFile{columns: all[0], rows: all[1:]}
	}

	return
}
