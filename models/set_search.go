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

	// first, we create the where conditions that are valid for both querying mysql
	// straight or querying sphinx first.
	var whereConds string
	var havingConds string
	if len(opts.Status) != 0 {
		whereConds = "ranked_status IN (" + sIntCommaSeparated(opts.Status) + ") "
	}
	if len(opts.Mode) != 0 {
		// This is a hack. Apparently, Sphinx does not support AND bitwise
		// operations in the WHERE clause, so we're placing that in the SELECT
		// clause and only making sure it's correct in this place.
		havingConds = " valid_set_modes = " + sm + " "
	}

	sets := make([]Set, 0, opts.Amount)
	setIDs := make([]int, 0, opts.Amount)
	// setMap is used when a query is given to make sure the results are kept in the correct
	// order given by sphinx.
	setMap := make(map[int]int, opts.Amount)
	// if Sphinx is used, limit will be cleared so that it's not used for the mysql query
	limit := fmt.Sprintf(" LIMIT %d, %d ", opts.Offset, opts.Amount)

	if opts.Query != "" {
		setIDsQuery := "SELECT id, set_modes & " + sm + " AS valid_set_modes FROM cg WHERE "

		// add filters to query
		// Yes. I know. Prepared statements. But Sphinx doesn't like them, so
		// bummer.
		setIDsQuery += "MATCH('" + mysqlStringReplacer.Replace(opts.Query) + "') "
		if whereConds != "" {
			setIDsQuery += "AND " + whereConds
		}
		if havingConds != "" {
			setIDsQuery += " AND " + havingConds
		}

		// set limit
		setIDsQuery += " ORDER BY WEIGHT() DESC, id DESC " + limit + " OPTION ranker=sph04, max_matches=20000 "
		limit = ""

		// fetch rows
		rows, err := searchDB.Query(setIDsQuery)
		if err != nil {
			return nil, err
		}

		// contains IDs of the sets we will retrieve
		for rows.Next() {
			var id int
			err = rows.Scan(&id, new(int))
			if err != nil {
				return nil, err
			}
			setIDs = append(setIDs, id)
			sets = sets[:len(sets)+1]
			setMap[id] = len(sets) - 1
		}

		// short path: there are no sets
		if len(sets) == 0 {
			return []Set{}, nil
		}

		whereConds = "id IN (" + sIntCommaSeparated(setIDs) + ")"
		havingConds = ""
	}

	if whereConds != "" {
		whereConds = "WHERE " + whereConds
	}
	if havingConds != "" {
		havingConds = " HAVING " + havingConds
	}
	setsQuery := "SELECT " + setFields + ", set_modes & " + sm + " AS valid_set_modes FROM sets " +
		whereConds + havingConds + " ORDER BY id DESC " + limit
	rows, err := db.Query(setsQuery)

	if err != nil {
		return nil, err
	}

	// find all beatmaps, but leave children aside for the moment.
	for rows.Next() {
		var s Set
		err = rows.Scan(
			&s.ID, &s.RankedStatus, &s.ApprovedDate, &s.LastUpdate, &s.LastChecked,
			&s.Artist, &s.Title, &s.Creator, &s.Source, &s.Tags, &s.HasVideo, &s.Genre,
			&s.Language, &s.Favourites, new(int),
		)
		if err != nil {
			return nil, err
		}
		// we get the position we should place s in from the setMap, this way we
		// keep the order of results as sphinx prefers.
		pos, ok := setMap[s.ID]
		if ok {
			sets[pos] = s
		} else {
			sets = append(sets, s)
			setIDs = append(setIDs, s.ID)
			setMap[s.ID] = len(sets) - 1
		}
	}

	if len(sets) == 0 {
		return sets, nil
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
