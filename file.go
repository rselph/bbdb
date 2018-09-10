package main

import (
	"encoding/csv"
	"io"
	"os"
)

type smartFile struct {
	f       io.Closer
	r       *csv.Reader
	columns []string
}

func openFile(fname string) (s *smartFile, err error) {
	r, err := os.Open(fname)
	if err != nil {
		return
	}

	s = &smartFile{}
	s.f = r
	s.r = csv.NewReader(r)
	s.r.ReuseRecord = true

	tmp, err := s.r.Read()
	if err != nil {
		s.f.Close()
		return
	}

	s.columns = make([]string, len(tmp))
	copy(s.columns, tmp)
	return
}

func (s *smartFile) Read() (row []string, err error) {
	row, err = s.r.Read()
	if row != nil {
		err = nil
	}

	return
}

func (s *smartFile) Close() {
	s.f.Close()
}
