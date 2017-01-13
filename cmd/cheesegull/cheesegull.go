// cheesegull is the main application binary of CheeseGull. Its intent is to
// function as an osu! beatmap mirror, fetching beatmaps from the osu! website
// and API, and saving those in a MySQL database. And doing it well.
package main

import (
	"fmt"
	"os"

	"runtime/debug"

	_ "github.com/go-sql-driver/mysql"
	"github.com/osuripple/cheesegull"
	"github.com/osuripple/cheesegull/app"
	"github.com/osuripple/cheesegull/downloader"
	"github.com/osuripple/cheesegull/providers/fileresolvers"
	"github.com/osuripple/cheesegull/providers/sql"
	osuapi "gopkg.in/thehowl/go-osuapi.v1"
	cli "gopkg.in/urfave/cli.v2"
)

// settings variables
var (
	mysqlDSN           string
	osuUsername        string
	osuPassword        string
	osuAPIKey          string
	disableStacktraces bool
	workers            uint
)

func main() {
	app := &cli.App{
		Name:      "CheeseGull",
		HelpName:  "cheesegull",
		Usage:     "an open source osu! beatmap mirror developed by Ripple, for Ripple",
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
				Name:        "osu-username",
				Usage:       "osu! username of the user who will download the beatmaps.",
				EnvVars:     []string{"OSU_USERNAME"},
				Destination: &osuUsername,
			},
			&cli.StringFlag{
				Name:        "osu-password",
				Usage:       "osu! password of the user who will download the beatmaps.",
				EnvVars:     []string{"OSU_PASSWORD"},
				Destination: &osuPassword,
			},
			&cli.StringFlag{
				Name:        "osu-api-key",
				Usage:       "osu! API key to fetch information about them.",
				EnvVars:     []string{"OSU_API_KEY"},
				Destination: &osuAPIKey,
			},
			&cli.BoolFlag{
				Name:        "disable-stacktraces",
				Usage:       "Disable stacktraces.",
				Value:       false,
				Destination: &disableStacktraces,
			},
			&cli.UintFlag{
				Name:        "workers",
				Usage:       "Number of workers downloading beatmaps.",
				Value:       4,
				Destination: &workers,
			},
		},
	}

	app.Run(os.Args)
}

func execute(c *cli.Context) error {
	fmt.Println("CheeseGull", cheesegull.Version)

	// Set up various components of the application.
	prov, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		return err
	}
	d, err := downloader.LogIn(osuUsername, osuPassword)
	if err != nil {
		return err
	}
	fp := fileresolvers.FileSystem{}
	api := osuapi.NewClient(osuAPIKey)
	if err := api.Test(); err != nil {
		return err
	}

	a := &app.App{
		Downloader: d,
		Service:    prov,
		FileResolver: fp,
		Source:     api,
		ErrorHandler: func(err error) {
			fmt.Println(err)
			if !disableStacktraces {
				fmt.Println(string(debug.Stack()))
			}
		},
	}

	fmt.Println("successfully initialised CheeseGull. Starting...")

	return a.Start(int(workers))
}
