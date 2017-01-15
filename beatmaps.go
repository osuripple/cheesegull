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

// SearchOptions are the various options for searching on the CheeseGull API.
type SearchOptions struct {
	// If len is 0, then it should be treated as if all statuses are good.
	Status []int
	// Query is what we're looking for.
	Query string
	// Game mode to which limit the results. If len is 0, it means all modes
	// are ok.
	Mode []int
	// Pagination options.
	Offset int
	Amount int // ... of results to return
}

// These are the sorting systems for ChunkOfSets
const (
	// SortLastChecked sorts by the date last checked, in descending order
	SortLastChecked = iota
)

// BeatmapService is a service connected to a database that can fetch
// information about beatmaps in the local DB.
type BeatmapService interface {
	Beatmaps(...int) ([]Beatmap, error)
	BeatmapSets(...int) ([]BeatmapSet, error)
	ChunkOfSets(offset, chunk, sortSystem int) ([]BeatmapSet, error)
	HighestBeatmapSetID() (int, error)
	SearchSets(SearchOptions) ([]BeatmapSet, error)
	CreateSet(BeatmapSet) error
	CreateBeatmaps(...Beatmap) error
}

// InheritFromOsuSet inherits properties from a slice of osuapi.Beatmaps.
// If recursive is enabled, ChildrenBeatmaps2 will be cleared and filled with
// beatmaps created with Beatmap.InheritFromOsuBeatmap.
func (b *BeatmapSet) InheritFromOsuSet(bms []osuapi.Beatmap, recursive bool) {
	if len(bms) == 0 {
		return
	}
	base := bms[0]

	b.ApprovedDate = time.Time(base.ApprovedDate)
	b.Artist = base.Artist
	b.Creator = base.Creator
	b.Favourites = base.FavouriteCount
	b.Genre = int(base.Genre)
	b.Language = int(base.Language)
	b.LastUpdate = time.Time(base.LastUpdate)
	b.RankedStatus = int(base.Approved)
	b.SetID = int(base.BeatmapSetID)
	b.Source = base.Source
	b.Tags = base.Tags
	b.Title = base.Title

	if recursive {
		b.ChildrenBeatmaps = make([]int, len(bms))
		b.ChildrenBeatmaps2 = make([]Beatmap, len(bms))

		for i, bm := range bms {
			b.ChildrenBeatmaps[i] = bm.BeatmapID
			newB := new(Beatmap)
			newB.InheritFromOsuBeatmap(bm)
			b.ChildrenBeatmaps2[i] = *newB
		}
	}
}

// InheritFromOsuBeatmap fills Beatmap with information that can be obtained
// by an osuapi.Beatmap.
func (b *Beatmap) InheritFromOsuBeatmap(bm osuapi.Beatmap) {
	b.AR = float32(bm.ApproachRate)
	b.BeatmapID = bm.BeatmapID
	b.BPM = bm.BPM
	b.CS = float32(bm.CircleSize)
	b.DifficultyRating = bm.DifficultyRating
	b.DiffName = bm.DiffName
	b.FileMD5 = bm.FileMD5
	b.HitLength = bm.HitLength
	b.HP = float32(bm.HPDrain)
	b.MaxCombo = bm.MaxCombo
	b.Mode = int(bm.Mode)
	b.OD = float32(bm.OverallDifficulty)
	b.ParentSetID = bm.BeatmapSetID
	b.Passcount = bm.Passcount
	b.Playcount = bm.Playcount
	b.TotalLength = bm.TotalLength
}
