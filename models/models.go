// Package models contains everything that is needed to interface to the
// database CheeseGull is using.
package models

import (
	"database/sql"
)

//go:generate go run migrations_gen.go

// RunMigrations brings the database up to date following the migrations.
func RunMigrations(db *sql.DB) error {
	var version int
	var _b []byte
	err := db.QueryRow("SHOW TABLES LIKE 'db_version'").Scan(&_b)
	switch err {
	case nil:
		// fetch version from db
		err = db.QueryRow("SELECT version FROM db_version").Scan(&version)
		if err != nil {
			return err
		}
	case sql.ErrNoRows:
		_, err = db.Exec("CREATE TABLE db_version(version INT NOT NULL)")
		if err != nil {
			return err
		}
		_, err = db.Exec("INSERT INTO db_version(version) VALUES ('-1')")
		if err != nil {
			return err
		}
		version = -1
	default:
		return err
	}

	for {
		version++
		if version >= len(migrations) {
			version--
			db.Exec("UPDATE db_version SET version = ?", version)
			return nil
		}

		s := migrations[version]
		_, err = db.Exec(s)
		if err != nil {
			return err
		}
	}
}
