// Package dbmirror is a package to create a database which is almost exactly
// the same as osu!'s beatmap database.
package dbmirror

import (
	"database/sql"
	"log"
	"os"
	"time"

	raven "github.com/getsentry/raven-go"
	"github.com/osuripple/cheesegull/models"
	osuapi "github.com/thehowl/go-osuapi"
)

const (
	// NewBatchEvery is the amount of time that will elapse between one batch
	// of requests and another.
	NewBatchEvery = time.Minute
	// PerBatch is the amount of requests and updates every batch contains.
	PerBatch = 100
	// SetUpdaterWorkers is the number of goroutines which should take care of
	// new batches. Keep in mind that this will be the number of maximum
	// concurrent connections to the osu! API.
	SetUpdaterWorkers = PerBatch / 20
)

// hasVideo checks whether a beatmap set has a video.
var hasVideo func(set int) (bool, error)

// SetHasVideo sets the hasVideo function to the one passed.
func SetHasVideo(f func(int) (bool, error)) {
	if f == nil {
		return
	}
	hasVideo = f
}

func createChildrenBeatmaps(bms []osuapi.Beatmap) []models.Beatmap {
	cgBms := make([]models.Beatmap, len(bms))
	for idx, bm := range bms {
		cgBms[idx] = models.Beatmap{
			ID:               bm.BeatmapID,
			ParentSetID:      bm.BeatmapSetID,
			DiffName:         bm.DiffName,
			FileMD5:          bm.FileMD5,
			Mode:             int(bm.Mode),
			BPM:              bm.BPM,
			AR:               float32(bm.ApproachRate),
			OD:               float32(bm.OverallDifficulty),
			CS:               float32(bm.CircleSize),
			HP:               float32(bm.HPDrain),
			TotalLength:      bm.TotalLength,
			HitLength:        bm.HitLength,
			Playcount:        bm.Playcount,
			Passcount:        bm.Passcount,
			MaxCombo:         bm.MaxCombo,
			DifficultyRating: bm.DifficultyRating,
		}
	}
	return cgBms
}

func setFromOsuAPIBeatmap(b osuapi.Beatmap) models.Set {
	return models.Set{
		ID:           b.BeatmapSetID,
		RankedStatus: int(b.Approved),
		ApprovedDate: time.Time(b.ApprovedDate),
		LastUpdate:   time.Time(b.LastUpdate),
		LastChecked:  time.Now(),
		Artist:       b.Artist,
		Title:        b.Title,
		Creator:      b.Creator,
		Source:       b.Source,
		Tags:         b.Tags,
		Genre:        int(b.Genre),
		Language:     int(b.Language),
		Favourites:   b.FavouriteCount,
	}
}

func updateSet(c *osuapi.Client, db *sql.DB, set models.Set) error {
	var (
		err error
		bms []osuapi.Beatmap
	)
	for i := 0; i < 5; i++ {
		bms, err = c.GetBeatmaps(osuapi.GetBeatmapsOpts{
			BeatmapSetID: set.ID,
		})
		if err == nil {
			break
		}
		if i >= 5 {
			return err
		}
	}
	if len(bms) == 0 {
		// set has been deleted from osu!, so we do the same thing
		return models.DeleteSet(db, set.ID)
	}

	// create the new set based on the information we can obtain from the
	// first beatmap's information
	var x = bms[0]
	updated := !time.Time(x.LastUpdate).Equal(set.LastUpdate)
	oldHasVideo := set.HasVideo
	set = setFromOsuAPIBeatmap(x)
	set.HasVideo = oldHasVideo
	set.ChildrenBeatmaps = createChildrenBeatmaps(bms)
	if updated {
		// if it has been updated, video might have been added or removed
		// so we need to check for it
		set.HasVideo, err = hasVideo(x.BeatmapSetID)
		if err != nil {
			return err
		}
	}

	return models.CreateSet(db, set)
}

// By making the buffer the same size of the batch, we can be sure that all
// sets from the previous batch will have completed by the time we finish
// pushing all the beatmaps to the queue.
var setQueue = make(chan models.Set, PerBatch)

// setUpdater is a function to be run as a goroutine, that receives sets
// from setQueue and brings the information in the database up-to-date for that
// set.
func setUpdater(c *osuapi.Client, db *sql.DB) {
	for set := range setQueue {
		err := updateSet(c, db, set)
		if err != nil {
			logError(err)
		}
	}
}

// StartSetUpdater does batch updates for the beatmaps in the database,
// employing goroutines to fetch the data from the osu! API and then write it to
// the database.
func StartSetUpdater(c *osuapi.Client, db *sql.DB) {
	for i := 0; i < SetUpdaterWorkers; i++ {
		go setUpdater(c, db)
	}
	for {
		sets, err := models.FetchSetsForBatchUpdate(db, PerBatch)
		if err != nil {
			logError(err)
			time.Sleep(NewBatchEvery)
			continue
		}
		for _, set := range sets {
			setQueue <- set
		}
		if len(sets) > 0 {
			log.Printf("[U] Updating sets, oldest LastChecked %v, newest %v, total length %d",
				sets[0].LastChecked,
				sets[len(sets)-1].LastChecked,
				len(sets),
			)
		}
		time.Sleep(NewBatchEvery)
	}
}

var envSentryDSN = os.Getenv("SENTRY_DSN")

// logError attempts to log an error to Sentry, as well as stdout.
func logError(err error) {
	if err == nil {
		return
	}
	if envSentryDSN != "" {
		raven.CaptureError(err, nil)
	}
	log.Println(err)
}
