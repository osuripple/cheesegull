// Package metadata handles API request that search for metadata regarding osu!
// beatmaps.
package metadata

import (
	"strconv"
	"strings"

	"github.com/osuripple/cheesegull/api"
	"github.com/osuripple/cheesegull/models"
)

// Beatmap handles requests to retrieve single beatmaps.
func Beatmap(c *api.Context) {
	id, _ := strconv.Atoi(strings.TrimSuffix(c.Param("id"), ".json"))
	if id == 0 {
		c.WriteJSON(404, nil)
		return
	}

	bms, err := models.FetchBeatmaps(c.DB, id)
	if err != nil {
		c.Err(err)
		c.WriteJSON(500, nil)
		return
	}
	if len(bms) == 0 {
		c.WriteJSON(404, nil)
		return
	}

	c.WriteJSON(200, bms[0])
}

// Set handles requests to retrieve single beatmap sets.
func Set(c *api.Context) {
	id, _ := strconv.Atoi(strings.TrimSuffix(c.Param("id"), ".json"))
	if id == 0 {
		c.WriteJSON(404, nil)
		return
	}

	set, err := models.FetchSet(c.DB, id, true)
	if err != nil {
		c.Err(err)
		c.WriteJSON(500, nil)
		return
	}
	if set == nil {
		c.WriteJSON(404, nil)
		return
	}

	c.WriteJSON(200, set)
}

func mustInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func mustPositive(i int) int {
	if i < 0 {
		return 0
	}
	return i
}

func intWithBounds(i, min, max, def int) int {
	if i == 0 {
		return def
	}
	if i < min {
		return min
	}
	if i > max {
		return max
	}
	return i
}

func sIntWithBounds(strs []string, min, max int) []int {
	sInt := make([]int, 0, len(strs))
	for _, s := range strs {
		i, err := strconv.Atoi(s)
		if err != nil || i < min || i > max {
			continue
		}
		sInt = append(sInt, i)
	}
	return sInt
}

// Search does a search on the sets available in the database.
func Search(c *api.Context) {
	query := c.Request.URL.Query()
	sets, err := models.SearchSets(c.DB, c.SearchDB, models.SearchOptions{
		Status: sIntWithBounds(query["status"], -2, 4),
		Query:  query.Get("query"),
		Mode:   sIntWithBounds(query["mode"], 0, 3),

		Amount: intWithBounds(mustInt(query.Get("amount")), 1, 100, 50),
		Offset: mustPositive(mustInt(query.Get("offset"))),
	})
	if err != nil {
		c.Err(err)
		c.WriteJSON(500, nil)
		return
	}

	c.WriteJSON(200, sets)
}

func init() {
	api.GET("/api/b/:id", Beatmap)
	api.GET("/b/:id", Beatmap)
	api.GET("/api/s/:id", Set)
	api.GET("/s/:id", Set)

	api.GET("/api/search", Search)
}
