package api

import (
	"net/http"
	"strconv"

	"github.com/osuripple/cheesegull"
	"github.com/osuripple/cheesegull/http/ctx"
)

type rbResponse struct {
	_base
	BeatmapSet *cheesegull.BeatmapSet
}

// RequestBeatmap handles requests to update a certain beatmap.
func RequestBeatmap(w http.ResponseWriter, r *http.Request, c *ctx.Context) {
	resp := rbResponse{}

	i, _ := strconv.Atoi(r.PostFormValue("set_id"))
	if i == 0 {
		resp.Message = "set_id was not passed or is not an int"
		j(w, 400, resp)
		return
	}

	l5m, err := c.Logging.UpdateInLast5Minutes(i)
	if err != nil {
		c.HandleError(err)
		j(w, 200, aec)
		return
	}
	if l5m {
		resp.Message = "that beatmap was already requested in the last 5 minutes!"
		j(w, 200, resp)
		return
	}

	err = c.Communication.SendBeatmapRequest(i)
	if err != nil {
		c.HandleError(err)
		j(w, 500, aec)
		return
	}

	resp.Ok = true
	bms, err := c.BeatmapService.BeatmapSets(i)
	if err != nil {
		c.HandleError(err)
	}
	if len(bms) > 0 {
		resp.BeatmapSet = &bms[0]
	}

	j(w, 200, resp)
}

type searchResponse struct {
	_base
	Sets []cheesegull.BeatmapSet
}

// Search searches in the beatmaps.
func Search(w http.ResponseWriter, r *http.Request, c *ctx.Context) {
	amt := c.QueryIntDefault("amount", 50)
	if amt < 0 {
		amt = 0
	}
	if amt > 100 {
		amt = 100
	}
	sets, err := c.BeatmapService.SearchSets(cheesegull.SearchOptions{
		Status: c.QueryIntMultiple("status"),
		Mode:   c.QueryIntMultiple("mode"),
		Query:  r.URL.Query().Get("query"),
		Amount: amt,
		Offset: c.QueryInt("offset"),
	})
	if err != nil {
		c.HandleError(err)
		j(w, 500, aec)
		return
	}
	resp := searchResponse{}
	resp.Ok = true
	resp.Sets = sets
	j(w, 200, resp)
}
