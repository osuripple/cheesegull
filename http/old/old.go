package old

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/osuripple/cheesegull"
	"github.com/osuripple/cheesegull/http/ctx"
)

// calls for the old API of mirror

// BeatmapSet returns a beatmapset from the database. /s/:id
func BeatmapSet(w http.ResponseWriter, r *http.Request, c *ctx.Context) {
	var s *cheesegull.BeatmapSet
	defer func() {
		j(s, w)
	}()

	w.Header().Add("Content-Type", "application/json; charset=UTF-8")

	i, _ := strconv.Atoi(strings.TrimSuffix(c.Param("id"), ".json"))
	if i == 0 {
		w.WriteHeader(404)
		return
	}

	sets, err := c.BeatmapService.BeatmapSets(i)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}
	if len(sets) == 0 {
		w.WriteHeader(404)
		return
	}
	s = &sets[0]
}

// Beatmap returns a beatmap from the database. /b/:id
func Beatmap(w http.ResponseWriter, r *http.Request, c *ctx.Context) {
	var b *cheesegull.Beatmap
	defer func() {
		j(b, w)
	}()

	w.Header().Add("Content-Type", "application/json; charset=UTF-8")

	i, _ := strconv.Atoi(strings.TrimSuffix(c.Param("id"), ".json"))
	if i == 0 {
		w.WriteHeader(404)
		return
	}

	beatmaps, err := c.BeatmapService.Beatmaps(i)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}
	if len(beatmaps) == 0 {
		w.WriteHeader(404)
		return
	}
	b = &beatmaps[0]
}

// IndexMD5 returns a fake MD5 hash of index.json, which is no more functional
// with cheesegull.
func IndexMD5(w http.ResponseWriter, r *http.Request, c *ctx.Context) {
	w.Header().Add("Content-Type", "text/plain; charset=UTF-8")

	// hash of "[]"
	const h = "d751713988987e9331980363e24189ce"
	w.Write([]byte(h))
}

// IndexJSON returns a fake index.json, which is no more functional
// with cheesegull.
func IndexJSON(w http.ResponseWriter, r *http.Request, c *ctx.Context) {
	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	w.Write([]byte("[]"))
}

// j writes JSON to the response writer
func j(obj interface{}, w http.ResponseWriter) {
	err := json.NewEncoder(w).Encode(obj)
	if err != nil {
		fmt.Println(err)
	}
}
