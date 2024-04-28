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
	osuAPIKey = kingpin.Flag("api-key", "osu! API key").Short('k').Envar("OSU_API_KEY").String()

	osuUsername = kingpin.Flag("osu-username", "osu! username (for downloading and fetching whether a beatmap has a video)").Short('u').Envar("OSU_USERNAME").String()
	osuPassword = kingpin.Flag("osu-password", "osu! password (for downloading and fetching whether a beatmap has a video)").Short('p').Envar("OSU_PASSWORD").String()

	beatconnectToken = kingpin.Flag("beatconnect-token", "beatconnect token. if provided, will use beatconnect rather than osu! website for downloading beatmaps").Envar("BEATCONNECT_TOKEN").String()

	allowUnranked = kingpin.Flag("allow-unranked", "Allow unranked beatmaps to be downloaded").Envar("ALLOW_UNRANKED").Default("false").Bool()

	mysqlDSN     = kingpin.Flag("mysql-dsn", "DSN of MySQL").Short('m').Default("root@/cheesegull").Envar("MYSQL_DSN").String()
	searchDSN    = kingpin.Flag("search-dsn", searchDSNDocs).Default("root@tcp(127.0.0.1:9306)/cheesegull").Envar("SEARCH_DSN").String()
	httpAddr     = kingpin.Flag("http-addr", "Address on which to take HTTP requests.").Short('a').Default("127.0.0.1:62011").Envar("HTTP_ADDR").String()
	maxDisk      = kingpin.Flag("max-disk", "Maximum number of GB used by beatmap cache.").Default("10").Envar("MAXIMUM_DISK").Float64()
	removeNonZip = kingpin.Flag("remove-non-zip", "Remove non-zip files.").Default("false").Bool()
	fckcfAddr    = kingpin.Flag("fckcf-addr", "fckcf http address").Envar("FCKCF_ADDR").String()
	cgbinPath    = kingpin.Flag("cgbin-path", "cgbin.db file path").Default("cgbin.db").Envar("CGBIN_PATH").String()
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
	house := housekeeper.New(*cgbinPath)
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
	var downloaderClient downloader.Client
	if *beatconnectToken != "" {
		fmt.Println("Using beatconnect")
		downloaderClient = downloader.NewBeatConnectClient(*beatconnectToken)
	} else {
		fmt.Println("Using osu! website")

		var reqPreparer downloader.LogInRequestPreparer
		if *fckcfAddr == "" {
			// No fckck address provided, disable it.
			reqPreparer = &downloader.EmptyLogInRequestPreparer{}
		} else {
			// Fckcf address provided, use it as a proxy
			reqPreparer = &downloader.FckCf{Address: *fckcfAddr}
		}

		downloaderClient, err = downloader.NewOsuClient(*osuUsername, *osuPassword, reqPreparer)
		if err != nil {
			fmt.Println("Can't log in into osu!:", err)
			os.Exit(1)
		}
	}

	d := downloader.NewDownloader(downloaderClient)

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
	panic(http.ListenAndServe(*httpAddr, api.CreateHandler(db, db2, house, d, api.Options{
		AllowUnranked: *allowUnranked,
	})))
}
