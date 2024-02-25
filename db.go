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
	query := `insert or ignore into drive_stats (`
	query += strings.Join(columns, ", ")
	query += `) values (`
	query += strings.Join(c.positions[1:len(columns)+1], ", ")
	query += `)`

	params := make([]interface{}, len(columns))

	tx, err := c.db.Begin()
	if err != nil {
		return ins, debugErr(err)
	}
	prepared, err := tx.Prepare(query)
	if err == nil {
		ins = &inserter{tx: tx, query: prepared, params: params}
	}
	return ins, debugErr(err)
}

func (ins *inserter) putRow(values []string) (err error) {
	for i := range values {
		values[i] = strings.TrimSpace(values[i])
		if values[i] != "" {
			ins.params[i] = values[i]
		} else {
			ins.params[i] = nil
		}
	}

	_, err = ins.query.Exec(ins.params...)
	return debugErr(err)
}

func (ins *inserter) commit() (err error) {
	err = ins.query.Close()
	if err != nil {
		_ = ins.tx.Rollback()
		return debugErr(err)
	}

	return debugErr(ins.tx.Commit())
}

func (ins *inserter) rollback() error {
	_ = ins.query.Close()
	return debugErr(ins.tx.Rollback())
}
