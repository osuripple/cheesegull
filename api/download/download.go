// Package download handles the API call to download an osu! beatmap set.
package download

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/osuripple/cheesegull/api"
	"github.com/osuripple/cheesegull/downloader"
	"github.com/osuripple/cheesegull/housekeeper"
	"github.com/osuripple/cheesegull/models"
)

func errorMessage(c *api.Context, code int, err string) {
	c.WriteHeader("Content-Type", "text/plain; charset=utf-8")
	c.Code(code)
	c.Write([]byte(err))
}

func existsQueryKey(c *api.Context, s string) bool {
	_, ok := c.Request.URL.Query()[s]
	return ok
}

// Download is the handler for a request to download a beatmap
func Download(c *api.Context) {
	// get the beatmap ID
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errorMessage(c, 400, "Malformed ID")
		return
	}

	// fetch beatmap set and make sure it exists.
	set, err := models.FetchSet(c.DB, id, false)
	if err != nil {
		c.Err(err)
		errorMessage(c, 500, "Could not fetch set")
		return
	}
	if set == nil {
		errorMessage(c, 404, "Set not found")
		return
	}
	if set.RankedStatus <= 0 {
		errorMessage(c, 406, "Unranked beatmap sets are currently not available for download, following a warning")
		return
	}

	// use novideo only when we are requested to get a beatmap having a video
	// and novideo is in the request
	noVideo := set.HasVideo && existsQueryKey(c, "novideo")

	cbm, shouldDownload := c.House.AcquireBeatmap(&housekeeper.CachedBeatmap{
		ID:         id,
		NoVideo:    noVideo,
		LastUpdate: set.LastUpdate,
	})

	if shouldDownload {
		err := downloadBeatmap(c.DLClient, cbm, c.House)
		if err != nil {
			c.Err(err)
			errorMessage(c, 500, "Internal error")
			return
		}
	} else {
		cbm.MustBeDownloaded()
	}

	cbm.SetLastRequested(time.Now())

	if cbm.FileSize() == 0 {
		errorMessage(c, 504, "The beatmap could not be downloaded (probably got deleted from the osu! website)")
		return
	}

	f, err := cbm.File()
	if err != nil {
		c.Err(err)
		errorMessage(c, 500, "Internal error")
		return
	}
	defer f.Close()

	c.WriteHeader("Content-Type", "application/octet-stream")
	c.WriteHeader("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fmt.Sprintf("%d %s - %s.osz", set.ID, set.Artist, set.Title)))
	c.WriteHeader("Content-Length", strconv.FormatUint(uint64(cbm.FileSize()), 10))
	c.Code(200)

	_, err = io.Copy(c, f)
	if err != nil {
		c.Err(err)
	}
}

func downloadBeatmap(c *downloader.Client, b *housekeeper.CachedBeatmap, house *housekeeper.House) error {
	log.Println("[⬇️]", b.String())

	var fileSize uint64
	defer func() {
		// We need to wrap this inside a function because this way the arguments
		// to DownloadCompleted are actually evaluated during the defer call.
		b.DownloadCompleted(fileSize, house)
	}()

	// Start downloading.
	r, err := c.Download(b.ID, b.NoVideo)
	if err != nil {
		if err == downloader.ErrNoRedirect {
			return nil
		}
		return err
	}
	defer r.Close()

	// open the file we will write the beatmap into
	f, err := b.CreateFile()
	if err != nil {
		return err
	}
	defer f.Close()

	fSizeRaw, err := io.Copy(f, r)
	fileSize = uint64(fSizeRaw)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	api.GET("/d/:id", Download)
}
