package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
	_ "github.com/go-sql-driver/mysql"
	osuapi "github.com/thehowl/go-osuapi"

	"github.com/osuripple/cheesegull/api"
	"github.com/osuripple/cheesegull/dbmirror"
	"github.com/osuripple/cheesegull/downloader"
	"github.com/osuripple/cheesegull/housekeeper"
	"github.com/osuripple/cheesegull/models"

	// Components of the API we want to use
	_ "github.com/osuripple/cheesegull/api/download"
	_ "github.com/osuripple/cheesegull/api/metadata"
)

const searchDSNDocs = `"DSN to use for fulltext searches. ` +
	`This should be a SphinxQL server. Follow the format of the MySQL DSN. ` +
	`This can be the same as MYSQL_DSN, and cheesegull will still run ` +
	`successfully, however what happens when search is tried is undefined ` +
	`behaviour and you should definetely bother to set it up (follow the README).`

var (
	osuAPIKey    = kingpin.Flag("api-key", "osu! API key").Short('k').Envar("OSU_API_KEY").String()
	osuUsername  = kingpin.Flag("osu-username", "osu! username (for downloading and fetching whether a beatmap has a video)").Short('u').Envar("OSU_USERNAME").String()
	osuPassword  = kingpin.Flag("osu-password", "osu! password (for downloading and fetching whether a beatmap has a video)").Short('p').Envar("OSU_PASSWORD").String()
	mysqlDSN     = kingpin.Flag("mysql-dsn", "DSN of MySQL").Short('m').Default("root@/cheesegull").Envar("MYSQL_DSN").String()
	searchDSN    = kingpin.Flag("search-dsn", searchDSNDocs).Default("root@tcp(127.0.0.1:9306)/cheesegull").Envar("SEARCH_DSN").String()
	httpAddr     = kingpin.Flag("http-addr", "Address on which to take HTTP requests.").Short('a').Default("127.0.0.1:62011").String()
	maxDisk      = kingpin.Flag("max-disk", "Maximum number of GB used by beatmap cache.").Default("10").Envar("MAXIMUM_DISK").Float64()
	removeNonZip = kingpin.Flag("remove-non-zip", "Remove non-zip files.").Default("false").Bool()
)

func addTimeParsing(dsn string) string {
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	dsn += sep + "parseTime=true&multiStatements=true"
	return dsn
}

func main() {
	kingpin.Parse()

	fmt.Println("CheeseGull", Version)
	api.Version = Version

	// set up housekeeper
	house := housekeeper.New()
	err := house.LoadState()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	house.MaxSize = uint64(float64(1024*1024*1024) * (*maxDisk))
	if *removeNonZip {
		house.RemoveNonZip()
		return
	}
	house.StartCleaner()

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

	// set up search
	db2, err := sql.Open("mysql", *searchDSN)
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
	go dbmirror.DiscoverEvery(c, db, time.Hour*6, time.Second*20)

	// create request handler
	panic(http.ListenAndServe(*httpAddr, api.CreateHandler(db, db2, house, d)))
}
