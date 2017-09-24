package models

import (
	"database/sql"
	"fmt"
	"time"
)

// Set represents a set of beatmaps usually sharing the same song.
type Set struct {
	ID               int `json:"SetID"`
	ChildrenBeatmaps []Beatmap
	RankedStatus     int
	ApprovedDate     time.Time
	LastUpdate       time.Time
	LastChecked      time.Time
	Artist           string
	Title            string
	Creator          string
	Source           string
	Tags             string
	HasVideo         bool
	Genre            int
	Language         int
	Favourites       int
}

// FetchSetsForBatchUpdate fetches limit sets from the database, sorted by
// LastChecked (asc, older first). Results are further filtered: if the set's
// RankedStatus is 3, 0 or -1 (qualified, pending or WIP), at least 30 minutes
// must have passed from LastChecked. For all other statuses, at least 4 days
// must have passed from LastChecked.
func FetchSetsForBatchUpdate(db *sql.DB, limit int) ([]Set, error) {
	n := time.Now()
	rows, err := db.Query(`
SELECT
	id, ranked_status, approved_date, last_update, last_checked,
	artist, title, creator, source, tags, has_video, genre,
	language, favourites
FROM sets
WHERE (ranked_status IN (3, 0, -1) AND last_checked <= ?) OR last_checked <= ?
ORDER BY last_checked ASC
LIMIT ?`,
		n.Add(-time.Minute*30),
		n.Add(-time.Hour*24*4),
		limit,
	)
	if err != nil {
		return nil, err
	}

	sets := make([]Set, 0, limit)
	for rows.Next() {
		var s Set
		err = rows.Scan(
			&s.ID, &s.RankedStatus, &s.ApprovedDate, &s.LastUpdate, &s.LastChecked,
			&s.Artist, &s.Title, &s.Creator, &s.Source, &s.Tags, &s.HasVideo, &s.Genre,
			&s.Language, &s.Favourites,
		)
		if err != nil {
			return nil, err
		}
		sets = append(sets, s)
	}

	return sets, nil
}

// DeleteSet deletes a set from the database, removing also its children
// beatmaps.
func DeleteSet(db *sql.DB, set int) error {
	_, err := db.Exec("DELETE FROM beatmaps WHERE parent_set_id = ?", set)
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM sets WHERE id = ?", set)
	return err
}

// createSetModes will generate the correct value for setModes, which is
// basically a bitwise enum containing the modes that are on a certain set.
func createSetModes(bms []Beatmap) (setModes uint8) {
	for _, bm := range bms {
		m := bm.Mode
		if m < 0 || m >= 4 {
			continue
		}
		setModes |= 1 << uint(m)
	}
	return setModes
}

// CreateSet creates (and updates) a beatmap set in the database.
func CreateSet(db *sql.DB, s Set) error {
	fmt.Println("CreateSet", s.ID)
	// delete existing set, if any.
	// This is mostly a lazy way to make sure updates work as well.
	err := DeleteSet(db, s.ID)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
INSERT INTO sets(
	id, ranked_status, approved_date, last_update, last_checked,
	artist, title, creator, source, tags, has_video, genre,
	language, favourites, set_modes
)
VALUES (
	?, ?, ?, ?, ?,
	?, ?, ?, ?, ?, ?, ?,
	?, ?, ?
)`, s.ID, s.RankedStatus, s.ApprovedDate, s.LastUpdate, s.LastChecked,
		s.Artist, s.Title, s.Creator, s.Source, s.Tags, s.HasVideo, s.Genre,
		s.Language, s.Favourites, createSetModes(s.ChildrenBeatmaps))
	if err != nil {
		return err
	}

	return CreateBeatmaps(db, s.ChildrenBeatmaps...)
}

// BiggestSetID retrieves the biggest set ID in the sets database. This is used
// by discovery to have a starting point from which to discover new beatmaps.
func BiggestSetID(db *sql.DB) (int, error) {
	var i int
	err := db.QueryRow("SELECT id FROM sets ORDER BY id DESC LIMIT 1").Scan(&i)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return i, err
}
