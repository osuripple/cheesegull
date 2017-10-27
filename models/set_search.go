package models

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
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

var mysqlStringReplacer = strings.NewReplacer(
	`\`, `\\`,
	`"`, `\"`,
	`'`, `\'`,
	"\x00", `\0`,
	"\n", `\n`,
	"\r", `\r`,
	"\x1a", `\Z`,
)

func sIntCommaSeparated(nums []int) string {
	b := bytes.Buffer{}
	for idx, num := range nums {
		b.WriteString(strconv.Itoa(num))
		if idx != len(nums)-1 {
			b.WriteString(", ")
		}
	}
	return b.String()
}

// SearchSets retrieves sets, filtering them using SearchOptions.
func SearchSets(db, searchDB *sql.DB, opts SearchOptions) ([]Set, error) {
	sm := strconv.Itoa(int(opts.setModes()))
	setIDsQuery := "SELECT id, set_modes & " + sm + " AS valid_set_modes FROM cg WHERE "

	// add filters to query
	// Yes. I know. Prepared statements. But Sphinx doesn't like them, so
	// bummer.
	setIDsQuery += "MATCH('" + mysqlStringReplacer.Replace(opts.Query) + "') "
	if len(opts.Status) != 0 {
		setIDsQuery += "AND ranked_status IN (" + sIntCommaSeparated(opts.Status) + ") "
	}
	if len(opts.Mode) != 0 {
		// This is a hack. Apparently, Sphinx does not support AND bitwise
		// operations in the WHERE clause, so we're placing that in the SELECT
		// clause and only making sure it's correct in this place.
		setIDsQuery += "AND valid_set_modes = " + sm + " "
	}

	// set limit
	setIDsQuery += fmt.Sprintf("ORDER BY WEIGHT() DESC, id DESC LIMIT %d, %d OPTION ranker=sph04", opts.Offset, opts.Amount)

	// fetch rows
	rows, err := searchDB.Query(setIDsQuery)
	if err != nil {
		return nil, err
	}

	// from the rows we will retrieve the IDs of all our sets.
	// we also pre-create the slices containing the sets we will fill later on
	// when we fetch the actual data.
	setIDs := make([]int, 0, opts.Amount)
	sets := make([]Set, 0, opts.Amount)
	// setMap, having an ID, points to a position of a set contained in sets.
	setMap := make(map[int]int, opts.Amount)
	for rows.Next() {
		var id int
		err = rows.Scan(&id, new(int))
		if err != nil {
			return nil, err
		}
		setIDs = append(setIDs, id)
		sets = append(sets, Set{})
		setMap[id] = len(sets) - 1
	}

	// short circuit: there are no sets
	if len(sets) == 0 {
		return []Set{}, nil
	}

	setsQuery := "SELECT " + setFields + " FROM sets WHERE id IN (" + inClause(len(setIDs)) + ")"
	args := sIntToSInterface(setIDs)

	rows, err = db.Query(setsQuery, args...)

	if err != nil {
		return nil, err
	}

	// find all beatmaps, but leave children aside for the moment.
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
		sets[setMap[s.ID]] = s
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
		parentSet, ok := setMap[b.ParentSetID]
		if !ok {
			continue
		}
		sets[parentSet].ChildrenBeatmaps = append(sets[parentSet].ChildrenBeatmaps, b)
	}

	return sets, nil
}
