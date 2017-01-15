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
	Downloader    cheesegull.BeatmapDownloader
	Source        cheesegull.BeatmapInfoSource
	Service       cheesegull.BeatmapService
	FileResolver  cheesegull.FileResolver
	Communication cheesegull.CommunicationService
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

	ch, err := a.Communication.BeatmapRequestsChannel()
	if err != nil {
		return err
	}
	go a.communicationSync(a.download, ch)

	offset := 0
	for {
		sets, err := a.Service.ChunkOfSets(offset, chunkSize, cheesegull.SortLastChecked)
		if err != nil {
			return err
		}

		// If the number of sets is == the chunkSize, it means we filled up an
		// entire chunk, so there's probably more and in the next db call we
		// should look into it.
		// If they are different, it means we finished the sets to go through
		// and thus we should discover new beatmaps.
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

		// check beatmaps are good and if they are, add them to the download
		// queue.
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
	i -= 5000
	if i < 1 {
		i = 1
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

func (a *App) communicationSync(dst chan<- cheesegull.BeatmapSet, src <-chan int) {
	for i := range src {
		fmt.Println("Adding", i)

		s := cheesegull.BeatmapSet{
			SetID: i,
		}
		b, err := a.CheckGood(&s)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// if it's false, it means that the beatmap could not be found
		if b {
			dst <- s
		}
	}
}
