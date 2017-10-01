package dbmirror

import (
	"database/sql"
	"time"

	"github.com/osuripple/cheesegull/models"
	osuapi "github.com/thehowl/go-osuapi"
)

// Discover discovers new beatmaps in the osu! database and adds them.
func Discover(c *osuapi.Client, db *sql.DB) error {
	id, err := models.BiggestSetID(db)
	if err != nil {
		return err
	}
	// failedAttempts is the number of consecutive failed attempts at fetching a
	// beatmap (by 'failed', in this case we mean exclusively when a request to
	// get_beatmaps returns no beatmaps)
	failedAttempts := 0
	for failedAttempts < 4096 {
		id++
		bms, err := c.GetBeatmaps(osuapi.GetBeatmapsOpts{
			BeatmapSetID: id,
		})
		if err != nil {
			return err
		}
		if len(bms) == 0 {
			failedAttempts++
			continue
		}
		failedAttempts = 0

		set := setFromOsuAPIBeatmap(bms[0])
		set.ChildrenBeatmaps = createChildrenBeatmaps(bms)
		set.HasVideo, err = hasVideo(bms[0].BeatmapSetID)
		if err != nil {
			return err
		}

		err = models.CreateSet(db, set)
		if err != nil {
			return err
		}
	}

	return nil
}

// DiscoverEvery runs Discover and waits for it to finish. If Discover returns
// an error, then it will wait errorWait before running Discover again. If
// Discover doesn't return any error, then it will wait successWait before
// running Discover again.
func DiscoverEvery(c *osuapi.Client, db *sql.DB, successWait, errorWait time.Duration) {
	for {
		err := Discover(c, db)
		if err == nil {
			time.Sleep(successWait)
		} else {
			logError(err)
			time.Sleep(errorWait)
		}
	}
}
