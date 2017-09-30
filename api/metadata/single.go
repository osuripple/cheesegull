// Package metadata handles API request that search for metadata regarding osu!
// beatmaps.
package metadata

import (
	"fmt"
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
		fmt.Println("Error fetching beatmap", err)
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
		fmt.Println("Error fetching beatmap", err)
		c.WriteJSON(500, nil)
		return
	}
	if set == nil {
		c.WriteJSON(404, nil)
		return
	}

	c.WriteJSON(200, set)
}

func init() {
	api.GET("/api/b/:id", Beatmap)
	api.GET("/b/:id", Beatmap)
	api.GET("/api/s/:id", Set)
	api.GET("/s/:id", Set)
}
