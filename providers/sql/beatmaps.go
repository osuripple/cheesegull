package sql

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/osuripple/cheesegull"
)

func (p *provider) Beatmaps(beatmapIDs ...int) ([]cheesegull.Beatmap, error) {
	query, params, err := sqlx.In("SELECT "+`
		id as beatmap_id, diff_name, file_md5, mode, bpm, ar, od, cs, hp,
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
	
	b.id as beatmap_id, b.diff_name, b.file_md5, b.mode, b.bpm, b.ar, b.od, b.cs, b.hp,
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
}

func (p *provider) ChunkOfSets(offset, chunk, sortSystem int) ([]cheesegull.BeatmapSet, error) {
	q := fmt.Sprintf(sortingSystems[sortSystem], offset, chunk)
	rows, err := p.db.Query(q)
	if err != nil {
		return nil, err
	}
	return p.sets(rows)
}

const searchQuery = "SELECT " + setsFields +
	"WHERE MATCH (s.artist, s.title, s.creator, s.source, s.tags) " +
	"AGAINST (? IN NATURAL LANGUAGE MODE)"

func (p *provider) SearchSets(q string) ([]cheesegull.BeatmapSet, error) {
	rows, err := p.db.Query(searchQuery, q)
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

			&b.BeatmapID, &b.DiffName, &b.FileMD5, &b.Mode, &b.BPM, &b.AR,
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
	_, err := p.db.Exec(`REPLACE INTO sets(
		set_id, ranked_status, approved_date, last_update, last_checked,
		artist, title, creator, source, tags, has_video, genre,
		language, favourites
	) VALUES (
		?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?,
		?, ?
	)`,
		s.SetID, s.RankedStatus, s.ApprovedDate, s.LastUpdate, s.LastChecked,
		s.Artist, s.Title, s.Creator, s.Source, s.Tags, s.HasVideo, s.Genre,
		s.Language, s.Favourites,
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
