package main

import (
	"database/sql"
	"strings"
)

type inserter struct {
	tx     *sql.Tx
	query  *sql.Stmt
	params []interface{}
}

func (c *db) prepare(columns []string) (ins *inserter, err error) {
	query := `insert into drive_stats (`
	query += strings.Join(columns, ", ")
	query += `) values (`
	query += strings.Repeat(`?, `, len(columns)-1)
	query += `?)`

	params := make([]interface{}, len(columns))

	tx, err := c.db.Begin()
	if err != nil {
		return
	}
	prepared, err := tx.Prepare(query)
	if err == nil {
		ins = &inserter{tx: tx, query: prepared, params: params}
	}
	return
}

func (ins *inserter) putRow(values []string) (err error) {
	for i := range values {
		if values[i] != "" {
			ins.params[i] = values[i]
		} else {
			ins.params[i] = nil
		}
	}

	_, err = ins.query.Exec(ins.params...)
	return
}

func (ins *inserter) commit() error {
	ins.query.Close()
	return ins.tx.Commit()
}

func (ins *inserter) rollback() error {
	ins.query.Close()
	return ins.tx.Rollback()
}
