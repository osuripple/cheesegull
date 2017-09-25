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

const beatmapFields = `
id, parent_set_id, diff_name, file_md5, mode, bpm,
ar, od, cs, hp, total_length, hit_length,
playcount, passcount, max_combo, difficulty_rating`

func readBeatmapsFromRows(rows *sql.Rows, capacity int) ([]Beatmap, error) {
	var err error
	bms := make([]Beatmap, 0, capacity)
	for rows.Next() {
		var b Beatmap
		err = rows.Scan(
			&b.ID, &b.ParentSetID, &b.DiffName, &b.FileMD5, &b.Mode, &b.BPM,
			&b.AR, &b.OD, &b.CS, &b.HP, &b.TotalLength, &b.HitLength,
			&b.Playcount, &b.Passcount, &b.MaxCombo, &b.DifficultyRating,
		)
		if err != nil {
			return nil, err
		}
		bms = append(bms, b)
	}

	return bms, rows.Err()
}

// FetchBeatmaps retrieves a list of beatmap knowing their IDs.
func FetchBeatmaps(db *sql.DB, ids ...int) ([]Beatmap, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	q := `SELECT ` + beatmapFields + ` FROM beatmaps WHERE id IN (`
	args := make([]interface{}, len(ids))

	for idx, id := range ids {
		if idx != 0 {
			q += ", "
		}
		q += "?"
		args[idx] = id
	}

	rows, err := db.Query(q+");", args...)
	if err != nil {
		return nil, err
	}

	return readBeatmapsFromRows(rows, len(ids))
}

// CreateBeatmaps adds beatmaps in the database.
func CreateBeatmaps(db *sql.DB, bms ...Beatmap) error {
	if len(bms) == 0 {
		return nil
	}

	q := `INSERT INTO beatmaps(` + beatmapFields + `) VALUES `
	const valuePlaceholder = `(
		?, ?, ?, ?, ?, ?,
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
			bm.ID, bm.ParentSetID, bm.DiffName, bm.FileMD5, bm.Mode, bm.BPM,
			bm.AR, bm.OD, bm.CS, bm.HP, bm.TotalLength, bm.HitLength,
			bm.Playcount, bm.Passcount, bm.MaxCombo, bm.DifficultyRating,
		)
	}

	_, err := db.Exec(q, args...)
	return err
}
