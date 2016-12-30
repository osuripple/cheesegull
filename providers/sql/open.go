// Package sql implements providers for cheesegull using SQL. This is tested
// with MySQL.
package sql

import (
	"net/url"

	"github.com/osuripple/cheesegull"
	"github.com/jmoiron/sqlx"
)

// Provided is a struct containing the services implemented by this package.
type Provided interface {
	cheesegull.BeatmapService
}

// provider is the actual provider containing all the methods.
type provider struct {
	db *sqlx.DB
}

// Open creates a new database
func Open(driver string, dsn string) (Provided, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("multiStatements", "true")
	q.Set("parseTime", "true")
	u.RawQuery = q.Encode()

	db, err := sqlx.Open(driver, u.String())
	if err != nil {
		return nil, err
	}
	db.MapperFunc(MapperFunc)
	err = AutoMigrateDB(db)
	if err != nil {
		return nil, err
	}
	p := &provider{db}
	return p, nil
}
