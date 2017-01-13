// Package http implements the CheeseGull API.
package http

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/osuripple/cheesegull"
	"github.com/osuripple/cheesegull/http/ctx"
	"github.com/osuripple/cheesegull/http/old"
)

// Options are the settings that can be passed to NewServer.
type Options struct {
	BeatmapService cheesegull.BeatmapService
}

// NewServer creates a new HTTP server for CheeseGull.
func NewServer(o Options) http.Handler {
	r := httprouter.New()

	r.GET("/s/:id", o.requestWrapper(old.BeatmapSet))
	r.GET("/b/:id", o.requestWrapper(old.Beatmap))
	r.GET("/index_md5.txt", o.requestWrapper(old.IndexMD5))
	r.GET("/index.json", o.requestWrapper(old.IndexJSON))
	// r.NotFound

	return r
}

func (o Options) requestWrapper(a func(w http.ResponseWriter, r *http.Request, c *ctx.Context)) func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a(w, r, &ctx.Context{
			Params:         p,
			BeatmapService: o.BeatmapService,
		})
	}
}
