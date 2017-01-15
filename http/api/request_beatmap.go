package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/osuripple/cheesegull"
	"github.com/osuripple/cheesegull/http/ctx"
)

type _base struct {
	Ok      bool
	Message string `json:",omitempty"`
}

const aec = "An error occurred."

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

	err := c.Communication.SendBeatmapRequest(i)
	if err != nil {
		c.HandleError(err)
		resp.Message = aec
		j(w, 500, resp)
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

// j writes JSON to the response writer
func j(w http.ResponseWriter, code int, obj interface{}) {
	err := json.NewEncoder(w).Encode(obj)
	if err != nil {
		fmt.Println(err)
	}
}
