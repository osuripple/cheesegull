// cheesegull is the main application binary of CheeseGull. Its intent is to
// function as an osu! beatmap mirror, fetching beatmaps from the osu! website
// and API, and saving those in a MySQL database. And doing it well.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/osuripple/cheesegull"
	"github.com/osuripple/cheesegull/providers/sql"
	pb "gopkg.in/cheggaaa/pb.v1"
	cli "gopkg.in/urfave/cli.v2"
)

var mysqlDSN string

func main() {
	app := &cli.App{
		Name:      "cheesegull-migrate-mirror",
		Usage:     "a tool for migrating from the old 'mirror' to the new cheesegull",
		Version:   "1.0.0",
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
		},
	}

	app.Run(os.Args)
}

var createSetChan = make(chan cheesegull.BeatmapSet, 100)
var errorList = make([]error, 0, 256)
var wg = new(sync.WaitGroup)
var p sql.Provided

func execute(c *cli.Context) error {
	var err error
	p, err = sql.Open("mysql", mysqlDSN)
	if err != nil {
		return err
	}

	for i := 0; i < 50; i++ {
		go thr()
		wg.Add(1)
	}

	files, err := ioutil.ReadDir("data/s/")
	if err != nil {
		return err
	}

	bar := pb.StartNew(len(files))
	bar.Start()

	for _, file := range files {
		bar.Increment()
		if file.IsDir() {
			continue
		}
		f, err := ioutil.ReadFile("data/s/" + file.Name())
		if err != nil {
			return err
		}

		var s cheesegull.BeatmapSet
		err = json.Unmarshal(f, &s)
		if err != nil {
			return err
		}

		for _, c := range s.ChildrenBeatmaps {
			data, err := ioutil.ReadFile(fmt.Sprintf("data/b/%d.json", c))
			if err != nil {
				errorList = append(errorList, err)
				continue
			}
			var b cheesegull.Beatmap
			err = json.Unmarshal(data, &b)
			if err != nil {
				return err
			}
			s.ChildrenBeatmaps2 = append(s.ChildrenBeatmaps2, b)
		}

		createSetChan <- s
	}
	close(createSetChan)
	wg.Wait()

	bar.FinishPrint("Done!")

	if len(errorList) > 0 {
		fmt.Println("Errors:")
		for _, err := range errorList {
			fmt.Println(err)
		}
	}

	return nil
}

func thr() {
	for i := range createSetChan {
		err := p.CreateSet(i)
		if err != nil {
			fmt.Println(err)
		}
	}
	wg.Done()
}
