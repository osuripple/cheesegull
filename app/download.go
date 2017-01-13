package app

import (
	"fmt"
	"io"

	"github.com/osuripple/cheesegull"
	"github.com/osuripple/cheesegull/providers/fileplacers"
)

// Worker is an instance that will download beatmaps as soon as they arrive down
// the channel.
func (a *App) Worker() {
	defer func() {
		err := recover()
		if err, ok := err.(error); ok {
			a.handle(err)
			go a.Worker()
		}
	}()
	for set := range a.download {
		err := a.Update(set)
		if err != nil {
			a.handle(err)
		}
	}
}

// Update takes care of updating a beatmap to the latest version.
func (a *App) Update(s cheesegull.BeatmapSet) error {
	fmt.Println("Updating", s.SetID, "...")
	normal, noVideo, err := a.Downloader.Download(s.SetID)
	if err != nil {
		// In the case of ErrNoRedirect, we should simply stop downloading,
		// there's no need to return the error because it is known and common
		// and to be expected.
		if err == cheesegull.ErrNoRedirect {
			return nil
		}
		return err
	}

	if a.FilePlacer == nil {
		a.FilePlacer = fileplacers.FileSystem{}
	}

	s.HasVideo = noVideo != nil

	if normal != nil {
		w, err := a.FilePlacer.Create(s.SetID, false)
		defer w.Close()
		if err != nil {
			return err
		}
		io.Copy(w, normal)
	}
	if noVideo != nil {
		w, err := a.FilePlacer.Create(s.SetID, true)
		defer w.Close()
		if err != nil {
			return err
		}
		io.Copy(w, noVideo)
	}

	fmt.Println("Finished updating", s.SetID)

	// we need to update hasVideo
	return a.Service.CreateSet(s)
}
