package models

import (
	"database/sql"
	"fmt"
)

// SearchOptions are options that can be passed to SearchSets for filtering
// sets.
type SearchOptions struct {
	// If len is 0, then it should be treated as if all statuses are good.
	Status []int
	Query  string
	// Gamemodes to which limit the results. If len is 0, it means all modes
	// are ok.
	Mode []int

	// Pagination options.
	Offset int
	Amount int
}

func (o SearchOptions) setModes() (total uint8) {
	for _, m := range o.Mode {
		if m < 0 || m >= 4 {
			continue
		}
		total |= 1 << uint8(m)
	}
	return
}

// SearchSets retrieves sets, filtering them using SearchOptions.
func SearchSets(db *sql.DB, opts SearchOptions) ([]Set, error) {
	setsQuery := "SELECT " + setFields +
		", MATCH(artist, title, creator, source, tags) AGAINST (? IN NATURAL LANGUAGE MODE) AS relevance FROM sets WHERE 1 "
	args := []interface{}{opts.Query}

	// add filters to query
	if len(opts.Status) != 0 {
		setsQuery += "AND ranked_status IN (" + inClause(len(opts.Status)) + ") "
		args = append(args, sIntToSInterface(opts.Status)...)
	}
	if len(opts.Mode) != 0 {
		setsQuery += "AND (set_modes & ?) = ? "
		sm := opts.setModes()
		args = append(args, sm, sm)
	}

	// set order by
	if opts.Query == "" {
		setsQuery += "ORDER BY id DESC "
	} else {
		setsQuery += "AND MATCH(artist, title, creator, source, tags) AGAINST (? IN NATURAL LANGUAGE MODE) ORDER BY relevance DESC "
		args = append(args, opts.Query)
	}

	// set limit
	setsQuery += fmt.Sprintf("LIMIT %d, %d", opts.Offset, opts.Amount)

	// fetch rows
	rows, err := db.Query(setsQuery, args...)
	if err != nil {
		return nil, err
	}

	sets := make([]Set, 0, opts.Amount)
	// setIDs is used to make the IN statement later on. setMap is used for
	// finding the beatmap to which append the child.
	setIDs := make([]int, 0, opts.Amount)
	setMap := make(map[int]*Set, opts.Amount)

	// find all beatmaps, but leave children aside for the moment.
	for rows.Next() {
		var s Set
		var rel float64
		err = rows.Scan(
			&s.ID, &s.RankedStatus, &s.ApprovedDate, &s.LastUpdate, &s.LastChecked,
			&s.Artist, &s.Title, &s.Creator, &s.Source, &s.Tags, &s.HasVideo, &s.Genre,
			&s.Language, &s.Favourites, &rel,
		)
		if err != nil {
			return nil, err
		}
		sets = append(sets, s)
		setIDs = append(setIDs, s.ID)
		setMap[s.ID] = &sets[len(sets)-1]
	}

	if len(sets) == 0 {
		return []Set{}, nil
	}

	rows, err = db.Query(
		"SELECT "+beatmapFields+" FROM beatmaps WHERE parent_set_id IN ("+
			inClause(len(setIDs))+")",
		sIntToSInterface(setIDs)...,
	)
	if err != nil {
		return nil, err
	}

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
		parentSet := setMap[b.ParentSetID]
		if parentSet == nil {
			continue
		}
		parentSet.ChildrenBeatmaps = append(parentSet.ChildrenBeatmaps, b)
	}

	return sets, nil
}
