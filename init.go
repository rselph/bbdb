package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

const (
	schemaVersion = 1
)

type db struct {
	db        *sql.DB
	positions []string
}

func newDB(driver, fname string, wipe bool) (c *db, err error) {
	c = &db{}
	c.positions = make([]string, 1024)
	for i := range c.positions {
		c.positions[i] = fmt.Sprintf("$%d", i)
	}

	c.db, err = sql.Open(driver, fname)
	if err != nil {
		c = nil
		return c, debugErr(err)
	}

	err = c.setFlags(driver)
	if err != nil {
		c = nil
		return c, debugErr(err)
	}

	if wipe {
		err = c.dropAll()
		if err != nil {
			c = nil
			return c, debugErr(err)
		}
	}

	err = c.load()
	if err != nil {
		c = nil
		return c, debugErr(err)
	}

	err = c.checkAndOpen()
	if err != nil {
		c = nil
		return c, debugErr(err)
	}

	return
}

func (c *db) close() {
	_, _ = c.db.Exec(`UPDATE meta SET lastclose = $1;`, time.Now())
	_ = c.db.Close()
}

func (c *db) setFlags(driver string) (err error) {
	switch driver {
	case "sqlite3":
		_, err = c.db.Exec(`PRAGMA journal_mode=WAL;`)
	}
	return debugErr(err)
}

func (c *db) load() (err error) {
	var metaCount int
	row := c.db.QueryRow(`
SELECT count(*) from meta;
`)
	err = row.Scan(&metaCount)
	if err != nil || metaCount != 1 {
		// There's no meta table.  Assume blank database and initialize it.
		err = c.createTables()
	}

	return
}

func (c *db) createTables() (err error) {
	var tx *sql.Tx
	tx, err = c.db.Begin()
	if err != nil {
		return debugErr(err)
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
		} else {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.Exec(`
CREATE TABLE meta
(
    unique_ordinal int PRIMARY KEY DEFAULT 1,
    lastopen timestamp null default null,
    lastclose timestamp null default null,
    schema_version int
);
`)
	if err != nil {
		return debugErr(err)
	}

	_, err = tx.Exec(`
insert into meta (schema_version) VALUES ($1);
`,
		schemaVersion)
	if err != nil {
		return debugErr(err)
	}

	_, err = tx.Exec(`
CREATE TABLE drive_stats (
    date VARCHAR(32) NOT NULL,
    serial_number VARCHAR(128) NOT NULL,
    model VARCHAR(128) NOT NULL,
    capacity_bytes BIGINT NOT NULL,
    failure SMALLINT NOT NULL,
	datacenter VARCHAR(128) NULL,
	cluster_id VARCHAR(128) NULL,
	vault_id VARCHAR(128) NULL,
	pod_id VARCHAR(128) NULL,
	pod_slot_num VARCHAR(128) NULL,
	is_legacy_format SMALLINT NULL,
    PRIMARY KEY (date, model, serial_number)
    );
`)
	if err != nil {
		return debugErr(err)
	}

	for i := 1; i < 256; i++ {
		_, err = tx.Exec(fmt.Sprintf(`
	    	ALTER TABLE drive_stats ADD COLUMN smart_%d_raw BIGINT NULL;
    		ALTER TABLE drive_stats ADD COLUMN smart_%d_normalized BIGINT NULL;`, i, i))
		if err != nil {
			return debugErr(err)
		}
	}

	_, err = tx.Exec(`
CREATE INDEX model_index ON drive_stats (model);
CREATE INDEX failure_index ON drive_stats (failure);
`)
	if err != nil {
		return debugErr(err)
	}

	return
}

func (c *db) checkAndOpen() (err error) {
	var vers int
	row := c.db.QueryRow(`
SELECT schema_version from meta;
`)
	err = row.Scan(&vers)
	if err != nil {
		return debugErr(err)
	}
	if vers > schemaVersion {
		err = errors.New("database schema version too new")
		return debugErr(err)
	}

	_, err = c.db.Exec(`UPDATE meta SET lastopen = $1;`, time.Now())

	return debugErr(err)
}

func (c *db) finishLoad() (err error) {
	_, err = c.db.Exec(`
--
-- Create a view that has the number of drive days for each
-- model, which is simply the number of rows in drive_stats
-- for that model.
--
CREATE VIEW IF NOT EXISTS drive_days AS 
    SELECT model, count(*) AS drive_days 
    FROM drive_stats 
    GROUP BY model;

--
-- Create a view that has the number of failures for each model.
--
CREATE VIEW IF NOT EXISTS failures AS
    SELECT model, count(*) AS failures
    FROM drive_stats
    WHERE failure = 1
    GROUP BY model;

--
-- Join the views together and compute the annual failure rate.
-- "drive years" is computed by dividing the number of drive days
-- by 365, and then the annual failure rate is simply the number
-- of failures divided by the number of drive years.  The result
-- is multiplied by 100 to get a percentage.
--
CREATE VIEW IF NOT EXISTS failure_rates AS
    SELECT drive_days.model AS model,
           drive_days.drive_days AS drive_days,
           failures.failures AS failures, 
           (100.0 * failures) / (drive_days / 365.0) AS annual_failure_rate
    FROM drive_days, failures
    WHERE drive_days.model = failures.model
    ORDER BY model;
`)
	return debugErr(err)
}

func (c *db) dropAll() (err error) {
	_, err = c.db.Exec(`
drop table if exists meta;
drop table if exists drive_stats;
drop view if exists drive_days;
drop view if exists failures;
drop view if exists failure_rates;
`)
	return debugErr(err)
}
