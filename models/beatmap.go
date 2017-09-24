package models

import "database/sql"

// Beatmap represents a single beatmap (difficulty) on osu!.
type Beatmap struct {
	ID               int `json:"BeatmapID"`
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

// CreateBeatmaps adds beatmaps in the database.
func CreateBeatmaps(db *sql.DB, bms ...Beatmap) error {
	if len(bms) == 0 {
		return nil
	}

	q := `
INSERT INTO beatmaps(
	id, parent_set_id, diff_name, mode, bpm,
	ar, od, cs, hp, total_length, hit_length,
	playcount, passcount, max_combo, difficulty_rating
)
VALUES `
	const valuePlaceholder = `(
		?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?,
		?, ?, ?, ?
	)`

	args := make([]interface{}, 0, 15*4)
	for idx, bm := range bms {
		if idx != 0 {
			q += ", "
		}
		q += valuePlaceholder
		args = append(args,
			bm.ID, bm.ParentSetID, bm.DiffName, bm.Mode, bm.BPM,
			bm.AR, bm.OD, bm.CS, bm.HP, bm.TotalLength, bm.HitLength,
			bm.Playcount, bm.Passcount, bm.MaxCombo, bm.DifficultyRating,
		)
	}

	_, err := db.Exec(q, args...)
	return err
}
