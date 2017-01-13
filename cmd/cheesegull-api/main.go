// cheesegull is the main application binary of CheeseGull. Its intent is to
// function as an osu! beatmap mirror, fetching beatmaps from the osu! website
// and API, and saving those in a MySQL database. And doing it well.
package main

import (
	"fmt"
	nhttp "net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/osuripple/cheesegull"
	"github.com/osuripple/cheesegull/http"
	"github.com/osuripple/cheesegull/providers/sql"
	cli "gopkg.in/urfave/cli.v2"
)

// settings variables
var (
	mysqlDSN string
	port     string
)

func main() {
	app := &cli.App{
		Name:      "CheeseGull API",
		HelpName:  "cheesegull-api",
		Usage:     "the CheeseGull public API.",
		Version:   cheesegull.Version,
		Copyright: "(c) Ripple 2016-2017 under the MIT license",
		Authors: []*cli.Author{
			{"Morgan Bazalgette", "the@howl.moe"},
		},
		Action: execute,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "mysql-dsn",
				Usage:       "The DSN of the MySQL database to which to connect and use as the main data source.",
				EnvVars:     []string{"MYSQL_DSN"},
				Value:       "root@/cheesegull",
				Destination: &mysqlDSN,
			},
			&cli.StringFlag{
				Name:        "port",
				Aliases:     []string{"p"},
				Usage:       "Port on which to accept HTTP connections",
				EnvVars:     []string{"PORT"},
				Value:       ":62011",
				Destination: &port,
			},
		},
	}

	app.Run(os.Args)
}

func execute(c *cli.Context) error {
	fmt.Println("CheeseGull API", cheesegull.Version)

	prov, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		return err
	}

	serv := http.NewServer(http.Options{
		BeatmapService: prov,
	})

	return nhttp.ListenAndServe(port, serv)
}
