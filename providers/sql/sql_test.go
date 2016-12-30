package sql

import (
	"fmt"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	_ "github.com/go-sql-driver/mysql"
	"github.com/osuripple/cheesegull"
)

var p Provided

func TestOpen(t *testing.T) {
	var err error
	p, err = Open("mysql", "root@/cheesegull")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateSet(t *testing.T) {
	if p == nil {
		t.Skip()
	}
	err := p.CreateSet(cheesegull.BeatmapSet{
		SetID: 1337,
		ChildrenBeatmaps2: []cheesegull.Beatmap{
			{
				ParentSetID: 1337,
				BeatmapID:   15618,
				DiffName:    "Il risotto con le erbette DI MERDA VA BENE",
				FileMD5:     fmt.Sprintf("%x", "laksldkanskwpoqw"),
			},
			{
				ParentSetID: 1337,
				BeatmapID:   15619,
				DiffName:    "kapri pontu",
				FileMD5:     fmt.Sprintf("%x", "i cosi giustiiii"),
			},
		},
		Creator:     "Michelangelo",
		Artist:      "Dario Fo",
		Title:       "BADADIBIDIBIBADABODOBODODE",
		LastChecked: time.Now(),
		LastUpdate:  time.Now().Add(-time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestBeatmaps(t *testing.T) {
	if p == nil {
		t.Skip()
	}
	_, err := p.Beatmaps(591, 1902, 2451)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSearchSets(t *testing.T) {
	if p == nil {
		t.Skip()
	}
	f, err := p.SearchSets("dario")
	t.Log("\n" + spew.Sdump(f))
	if err != nil {
		t.Fatal(err)
	}
}
