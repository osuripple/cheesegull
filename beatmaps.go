package cheesegull

import (
	"time"

	"gopkg.in/thehowl/go-osuapi.v1"
)

// BeatmapInfoSource is a source able to provide information about multiple
// beatmaps. This will generally be a Client from go-osuapi, but also any mock
// client will fit.
type BeatmapInfoSource interface {
	GetBeatmaps(opts osuapi.GetBeatmapsOpts) ([]osuapi.Beatmap, error)
}

// BeatmapSet is a beatmapset containing a set of beatmaps.
type BeatmapSet struct {
	SetID             int
	ChildrenBeatmaps  []int
	ChildrenBeatmaps2 []Beatmap
	RankedStatus      int
	ApprovedDate      time.Time
	LastUpdate        time.Time
	LastChecked       time.Time
	Artist            string
	Title             string
	Creator           string
	Source            string
	Tags              string
	HasVideo          bool
	Genre             int
	Language          int
	Favourites        int
}

// Beatmap is a single beatmap, not an entire beatmapset.
type Beatmap struct {
	BeatmapID        int
	ParentSetID      int
	DiffName         string
	FileMD5          string
	Mode             int
	BPM              float64
	AR               float32
	OD               float32
	CS               float32
	HP               float32
	TotalLength      int
	HitLength        int
	Playcount        int
	Passcount        int
	MaxCombo         int
	DifficultyRating float64
}

// BeatmapService is a service connected to a database that can fetch
// information about beatmaps in the local DB.
type BeatmapService interface {
	Beatmaps(...int) ([]Beatmap, error)
	BeatmapSets(...int) ([]BeatmapSet, error)
	SearchSets(string) ([]BeatmapSet, error)
	CreateSet(BeatmapSet) error
	CreateBeatmaps(...Beatmap) error
}
