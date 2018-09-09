package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

const (
	schemaVersion = 1
)

type db struct {
	db *sql.DB
}

func newDB(fname string) (c *db, err error) {
	c = &db{}
	c.db, err = sql.Open("sqlite3", fname)
	if err != nil {
		c = nil
		return
	}

	err = c.load()
	if err != nil {
		c = nil
		return
	}

	return
}

func (c *db) load() (err error) {
	var metaCount int
	row := c.db.QueryRow(`
SELECT count(*) from sqlite_master where type = "table" and name = "meta";
`)
	err = row.Scan(&metaCount)
	if err != nil {
		return
	}
	if metaCount == 0 {
		// There's no meta table.  Assume blank database and initialize it.
		c.createTables()
	}

	return
}

func (c *db) createTables() (err error) {
	var tx *sql.Tx
	tx, err = c.db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec(`
CREATE TABLE meta
(
    "unique" int PRIMARY KEY DEFAULT 1,
    lastopen timestamp,
    lastclose timestamp,
    schema_version int
);
`)
	if err != nil {
		return
	}

	_, err = tx.Exec(`
insert into meta (schema_version) VALUES ($1);
`,
		schemaVersion)
	if err != nil {
		return
	}

	_, err = tx.Exec(`
CREATE TABLE drive_stats (
    date TEXT NOT NULL,
    serial_number TEXT NOT NULL,
    model TEXT NOT NULL,
    capacity_bytes INTEGER (8) NOT NULL,
    failure INTEGER (1) NOT NULL,
    PRIMARY KEY (date, model, serial_number)
    );
CREATE INDEX IF NOT EXISTS model_index ON drive_stats (model);
CREATE INDEX IF NOT EXISTS failure_index ON drive_stats (failure);
`)
	if err != nil {
		return
	}

	for i := 1; i < 256 ; i++ {
		_, err = tx.Exec(fmt.Sprintf(`
	    	ALTER TABLE drive_stats ADD COLUMN smart_%d_raw INTEGER;
    		ALTER TABLE drive_stats ADD COLUMN smart_%d_normalized INTEGER;
`, i, i))
		if err != nil {
			return
		}
	}

	_, err = tx.Exec(`
--
-- Create a view that has the number of drive days for each
-- model, which is simply the number of rows in drive_stats
-- for that model.
--
CREATE VIEW drive_days AS 
    SELECT model, count(*) AS drive_days 
    FROM drive_stats 
    GROUP BY model;

--
-- Create a view that has the number of failures for each model.
--
CREATE VIEW failures AS
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
CREATE VIEW failure_rates AS
    SELECT drive_days.model AS model,
           drive_days.drive_days AS drive_days,
           failures.failures AS failures, 
           (100.0 * failures) / (drive_days / 365.0) AS annual_failure_rate
    FROM drive_days, failures
    WHERE drive_days.model = failures.model
    ORDER BY model;

`)
	if err != nil {
		return
	}

	return
}
