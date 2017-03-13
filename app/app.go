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

const chunkSize = 5000

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

		if len(sets) == 0 {
			fmt.Println("Starting discoverNew")
			err := a.discoverNew()
			if err != nil {
				return err
			}
			// repeat from the beginning
			offset = 0
			continue
		}

		// check beatmaps are good and if they are, add them to the download
		// queue.
		for _, s := range sets {
			fmt.Println("Checking whether", s.SetID, "is good...")
			b, err := a.CheckGood(&s)
			if err != nil {
				a.handle(err)
				continue
			}

			if b {
				fmt.Println("Queueing", s.SetID)
				a.download <- s
			}
		}

		offset += chunkSize
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
