package sql

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/osuripple/cheesegull"
)

func (p *provider) Beatmaps(beatmapIDs ...int) ([]cheesegull.Beatmap, error) {
	query, params, err := sqlx.In("SELECT "+`
		id as beatmap_id, parent_id, diff_name, file_md5, mode, bpm, ar, od, cs, hp,
		total_length, hit_length, playcount, passcount, max_combo,
		difficulty_rating`+" FROM beatmaps WHERE id IN (?)", beatmapIDs)
	if err != nil {
		return nil, err
	}
	b := make([]cheesegull.Beatmap, 0, len(beatmapIDs))
	err = p.db.Select(&b, query, params...)
	return b, err
}

const setsFields = `
	s.set_id, s.ranked_status, s.approved_date, s.last_update, s.last_checked,
	s.artist, s.title, s.creator, s.source, s.tags, s.has_video, s.genre,
	s.language, s.favourites,

	b.id as beatmap_id, b.parent_id, b.diff_name, b.file_md5, b.mode, b.bpm, b.ar, b.od, b.cs, b.hp,
	b.total_length, b.hit_length, b.playcount, b.passcount, b.max_combo,
	b.difficulty_rating
FROM beatmaps b
INNER JOIN sets s ON b.parent_id = s.set_id
`

func (p *provider) BeatmapSets(sets ...int) ([]cheesegull.BeatmapSet, error) {
	q, a, err := sqlx.In("SELECT "+setsFields+" WHERE s.set_id IN (?)", sets)
	if err != nil {
		return nil, err
	}
	rows, err := p.db.Query(q, a...)
	if err != nil {
		return nil, err
	}
	return p.sets(rows)
}

var sortingSystems = [...]string{
	// SortLastChecked
	"SELECT " + setsFields + " ORDER BY s.last_checked DESC LIMIT %d, %d",
	// SortID
	"SELECT " + setsFields + " ORDER BY s.set_id DESC LIMIT %d, %d",
}

func (p *provider) ChunkOfSets(offset, chunk, sortSystem int) ([]cheesegull.BeatmapSet, error) {
	q := fmt.Sprintf(sortingSystems[sortSystem], offset, chunk)
	rows, err := p.db.Query(q)
	if err != nil {
		return nil, err
	}
	return p.sets(rows)
}

func _and(needAnd bool) string {
	if needAnd {
		return " AND "
	}
	return " WHERE "
}

// Alright, so since if we are asked to return x number of results we have to
// return x number of *sets*, what this does is: getting first of all all of the
// distinct beatmap set IDs of matching sets, then querying with an IN with the
// standard query done for calling the .sets method.
// See this: http://stackoverflow.com/a/16257924/5328069
// This is basically a clusterfuck, but it's the best way I could design it
// without dropping a single MS of speed.
func (p *provider) SearchSets(opts cheesegull.SearchOptions) ([]cheesegull.BeatmapSet, error) {
	queryBase := "SELECT DISTINCT set_id FROM sets "
	params := make([]interface{}, 0, 5)
	needAnd := false

	if len(opts.Mode) > 0 {
		queryBase += _and(needAnd) + "set_modes & ? = ?"
		mte := modesToEnum(opts.Mode)
		params = append(params, mte, mte)
		needAnd = true
	}
	if len(opts.Status) > 0 {
		queryBase += _and(needAnd) + "ranked_status IN (?)"
		params = append(params, opts.Status)
		needAnd = true
	}
	if opts.Query != "" {
		queryBase += _and(needAnd) +
			"MATCH (artist, title, creator, source, tags) AGAINST (? IN NATURAL LANGUAGE MODE)"
		params = append(params, opts.Query)
	}

	queryBase += fmt.Sprintf(" ORDER BY (MATCH (artist, title, creator, source, tags) AGAINST (? IN NATURAL LANGUAGE MODE)) DESC, set_id DESC LIMIT %d, %d", opts.Offset, opts.Amount)
	params = append(params, opts.Query)

	queryBase, params, err := sqlx.In(queryBase, params...)

	if err != nil {
		return nil, err
	}

	query := "SELECT" + setsFields + " INNER JOIN (" + queryBase + ") as iq ON s.set_id = iq.set_id"

	rows, err := p.db.Query(query, params...)
	if err != nil {
		return nil, err
	}
	return p.sets(rows)
}

func (p *provider) sets(rows *sql.Rows) ([]cheesegull.BeatmapSet, error) {
	var bms []cheesegull.BeatmapSet
RowLoop:
	for rows.Next() {
		var (
			b cheesegull.Beatmap
			s cheesegull.BeatmapSet
		)
		err := rows.Scan(
			&s.SetID, &s.RankedStatus, &s.ApprovedDate, &s.LastUpdate,
			&s.LastChecked, &s.Artist, &s.Title, &s.Creator, &s.Source, &s.Tags,
			&s.HasVideo, &s.Genre, &s.Language, &s.Favourites,

			&b.BeatmapID, &b.ParentSetID, &b.DiffName, &b.FileMD5, &b.Mode, &b.BPM, &b.AR,
			&b.OD, &b.CS, &b.HP, &b.TotalLength, &b.TotalLength, &b.Playcount,
			&b.Passcount, &b.MaxCombo, &b.DifficultyRating,
		)
		if err != nil {
			return nil, err
		}
		b.ParentSetID = s.SetID
		// We are looping because we must base ourselves on the assumption
		// that the DB returns rows randomly sorted.
		for i, s2 := range bms {
			if s2.SetID != s.SetID {
				continue
			}
			bms[i].ChildrenBeatmaps = append(bms[i].ChildrenBeatmaps, b.BeatmapID)
			bms[i].ChildrenBeatmaps2 = append(bms[i].ChildrenBeatmaps2, b)
			continue RowLoop
		}
		s.ChildrenBeatmaps = []int{b.BeatmapID}
		s.ChildrenBeatmaps2 = []cheesegull.Beatmap{b}
		bms = append(bms, s)
	}
	return bms, nil
}

func (p *provider) HighestBeatmapSetID() (i int, err error) {
	err = p.db.Get(&i, "SELECT set_id FROM sets ORDER BY set_id DESC LIMIT 1")
	if err == sql.ErrNoRows {
		err = nil
	}
	return
}

func (p *provider) CreateSet(s cheesegull.BeatmapSet) error {
	modes := make([]int, 0, len(s.ChildrenBeatmaps))
	for _, x := range s.ChildrenBeatmaps2 {
		modes = append(modes, x.Mode)
	}
	_, err := p.db.Exec(`INSERT INTO sets(
		set_id, ranked_status, approved_date, last_update, last_checked,
		artist, title, creator, source, tags, has_video, genre,
		language, favourites, set_modes
	) VALUES (
		?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?,
		?, ?, ?
	) ON DUPLICATE KEY UPDATE
		set_id = VALUES(set_id), ranked_status = VALUES(ranked_status),
		approved_date = VALUES(approved_date),
		last_update = VALUES(last_update), last_checked = VALUES(last_checked),
		artist = VALUES(artist), title = VALUES(title),
		creator = VALUES(creator), source = VALUES(source), tags = VALUES(tags),
		has_video = VALUES(has_video), genre = VALUES(genre),
		language = VALUES(language), favourites = VALUES(favourites),
		set_modes = VALUES(set_modes)
	`,
		s.SetID, s.RankedStatus, s.ApprovedDate, s.LastUpdate, s.LastChecked,
		s.Artist, s.Title, s.Creator, s.Source, s.Tags, s.HasVideo, s.Genre,
		s.Language, s.Favourites, modesToEnum(modes),
	)
	if len(s.ChildrenBeatmaps2) > 0 {
		err := p.CreateBeatmaps(s.ChildrenBeatmaps2...)
		if err != nil {
			return err
		}
	}
	return err
}

func (p *provider) CreateBeatmaps(bms ...cheesegull.Beatmap) error {
	if len(bms) == 0 {
		return nil
	}
	q := `REPLACE INTO beatmaps(
		id, parent_id, diff_name, file_md5, mode, bpm, ar, od, cs, hp,
		total_length, hit_length, playcount, passcount, max_combo,
		difficulty_rating
	) VALUES `
	// There are 15 fields, so yeah
	args := make([]interface{}, 0, 15*len(bms))
	for i, bm := range bms {
		q += "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
		if i != len(bms)-1 {
			q += ", "
		}
		args = append(args,
			bm.BeatmapID, bm.ParentSetID, bm.DiffName, bm.FileMD5, bm.Mode,
			bm.BPM, bm.AR, bm.OD, bm.CS, bm.HP, bm.TotalLength, bm.HitLength,
			bm.Playcount, bm.Passcount, bm.MaxCombo, bm.DifficultyRating,
		)
	}
	_, err := p.db.Exec(q, args...)
	return err
}
