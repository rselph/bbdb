package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type idTable struct {
	objNames []string
	table    string
	cache    map[string]interface{}

	db *sql.DB
}

func newIdTable(db *sql.DB, table string, objNames ...string) (idt *idTable, err error) {
	idt = &idTable{
		objNames: objNames,
		table:    table,
		cache:    make(map[string]interface{}),
		db:       db,
	}

	return
}

func (idt *idTable) getId(names ...string) (id int64, err error) {
	if len(names) != len(idt.objNames) {
		return 0, errors.New("wrong number of object names")
	}
	return
}

func (idt *idTable) getNames(id int64) (names []string, err error) {
	return
}

func (idt *idTable) getName(id int64) (name string, err error) {
	names, err := idt.getNames(id)
	if names != nil {
		name = names[0]
	}
	return
}

func (idt *idTable) createDBTable() (err error) {
	_, err = idt.db.Exec(fmt.Sprintf(`select coun(*) from %s;`, idt.table))
	if err == nil {
		return
	}

	_, err = idt.db.Exec(fmt.Sprintf(`
create table %s (
    id BIGINT NOT NULL,
    %s VARCHAR(128) NOT NULL,
    PRIMARY KEY (id)
);`, idt.table))
	if err != nil {
		return debugErr(err)
	}

	for _, col := range idt.objNames {
		_, err = idt.db.Exec(fmt.Sprintf(
			`ALTER TABLE %s ADD COLUMN %s VARCHAR(128) NOT NULL`, idt.table, col))
		if err != nil {
			return debugErr(err)
		}
	}
	_, err = idt.db.Exec(fmt.Sprintf(`CREATE UNIQUE INDEX %s_names ON %s (%s);`,
		idt.table, strings.Join(idt.objNames, ", ")))

	return
}
