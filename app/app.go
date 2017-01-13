// Package app contains the basic application logic and assembles together all
// the pieces of CheeseGull.
package app

import (
	"fmt"

	"github.com/osuripple/cheesegull"
)

// App is the basic struct of the application which contains all the methods
// to make it run smoothly.
type App struct {
	Downloader cheesegull.BeatmapDownloader
	Source     cheesegull.BeatmapInfoSource
	Service    cheesegull.BeatmapService
	FilePlacer cheesegull.FilePlacer
	// handles non-critical errors
	ErrorHandler func(err error)
	download     chan cheesegull.BeatmapSet
}

const chunkSize = 200

// Start starts the application as a whole. Start should never return, unless
// there's a critical error.
func (a *App) Start(n int) error {
	// initialisation
	a.download = make(chan cheesegull.BeatmapSet, 64)
	if n < 1 {
		n = 1
	}
	for i := 0; i < n; i++ {
		go a.Worker()
	}

	offset := 0
	for {
		sets, err := a.Service.ChunkOfSets(offset, chunkSize, cheesegull.SortLastChecked)
		if err != nil {
			return err
		}

		if len(sets) == chunkSize {
			offset += chunkSize
		} else {
			err := a.discoverNew()
			if err != nil {
				return err
			}
			// repeat from the beginning
			offset = 0
		}

		for _, s := range sets {
			b, err := a.CheckGood(&s)
			if err != nil {
				a.handle(err)
				continue
			}

			if b {
				a.download <- s
			}
		}
	}
}

// discoverNew is the part where we discover new beatmaps that are available to download.
func (a *App) discoverNew() error {
	i, err := a.Service.HighestBeatmapSetID()
	if err != nil {
		return err
	}
	// number of consecutive beatmaps not found
	var notFound int
	for {
		i++
		if notFound > 5000 {
			return nil
		}

		s := cheesegull.BeatmapSet{
			SetID: i,
		}
		b, err := a.CheckGood(&s)
		if err != nil {
			return err
		}

		if !b {
			notFound++
			continue
		}

		a.download <- s
	}
}

func (a *App) handle(err error) {
	if a.ErrorHandler == nil {
		fmt.Println(err)
		return
	}
	a.ErrorHandler(err)
}
