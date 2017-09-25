package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/osuripple/cheesegull/dbmirror"
	"github.com/osuripple/cheesegull/downloader"
	"github.com/osuripple/cheesegull/models"
	osuapi "github.com/thehowl/go-osuapi"
)

var (
	osuAPIKey   = kingpin.Flag("api-key", "osu! API key").Short('k').Envar("OSU_API_KEY").String()
	osuUsername = kingpin.Flag("osu-username", "osu! username (for downloading and fetching whether a beatmap has a video)").Short('u').Envar("OSU_USERNAME").String()
	osuPassword = kingpin.Flag("osu-password", "osu! password (for downloading and fetching whether a beatmap has a video)").Short('p').Envar("OSU_PASSWORD").String()
	mysqlDSN    = kingpin.Flag("mysql-dsn", "DSN of MySQL").Short('m').Default("root@/cheesegull").Envar("MYSQL_DSN").String()
)

func addTimeParsing(dsn string) string {
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	dsn += sep + "parseTime=true"
	return dsn
}

func main() {
	kingpin.Parse()

	fmt.Println("CheeseGull", Version)

	// set up osuapi client
	c := osuapi.NewClient(*osuAPIKey)

	// set up downloader
	d, err := downloader.LogIn(*osuUsername, *osuPassword)
	if err != nil {
		fmt.Println("Can't log in into osu!:", err)
		os.Exit(1)
	}
	dbmirror.SetHasVideo(d.HasVideo)

	// set up mysql
	db, err := sql.Open("mysql", addTimeParsing(*mysqlDSN))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// run mysql migrations
	err = models.RunMigrations(db)
	if err != nil {
		fmt.Println("Error running migrations", err)
	}

	// start running components of cheesegull
	go dbmirror.StartSetUpdater(c, db)
	dbmirror.DiscoverEvery(c, db, time.Minute*30, time.Second*20)
}
