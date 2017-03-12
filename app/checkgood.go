package app

import (
	"time"

	"github.com/osuripple/cheesegull"
	"gopkg.in/thehowl/go-osuapi.v1"
)

// CheckGood checks whether a beatmap is good or not to be downloaded.
func (a *App) CheckGood(s *cheesegull.BeatmapSet) (bool, error) {
	if !enoughTimePassed(*s) {
		return false, nil
	}

	var (
		beatmaps []osuapi.Beatmap
		err      error
		attempts int
	)
	for {
		beatmaps, err = a.Source.GetBeatmaps(osuapi.GetBeatmapsOpts{
			BeatmapSetID: s.SetID,
		})
		if err != nil {
			attempts++
			if attempts > 5 {
				return false, err
			}
			time.Sleep(3 * time.Second)
			continue
		}
		if len(beatmaps) == 0 {
			return false, nil
		}
		break
	}

	// Update beatmap
	s.InheritFromOsuSet(beatmaps, true)
	s.LastChecked = time.Now()
	err = a.Service.CreateSet(*s)

	return beatmaps[0].LastUpdate.GetTime().Equal(s.LastUpdate), nil
}

func enoughTimePassed(bm cheesegull.BeatmapSet) bool {
	var multiplier float64

	// should update immediately
	if bm.LastUpdate.IsZero() || bm.LastChecked.IsZero() {
		return true
	}

	switch bm.RankedStatus {
	// ranked, approved
	case 1, 2:
		multiplier = float64(1) / 4 // rarely checked
	// loved
	case 4:
		multiplier = float64(1) // isn't likely to change much, but still more than a ranked or approved
	// qualified
	case 3:
		multiplier = float64(8) // qualified must be checked very often so that we know if it's been ranked
	// pending, wip
	case 0, -1:
		multiplier = float64(4) // must be checked often because they can change quickly
	// graveyard
	case -2:
		multiplier = float64(1) / 6 // really unlikely to change
	}

	return time.Since(bm.LastChecked).Seconds() > (float64(time.Now().Unix()-bm.LastUpdate.Unix()) / (20 * multiplier))
}
