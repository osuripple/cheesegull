// Package http implements the CheeseGull API.
package http

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/osuripple/cheesegull"
	"github.com/osuripple/cheesegull/http/api"
	"github.com/osuripple/cheesegull/http/ctx"
	"github.com/osuripple/cheesegull/http/old"
)

// Options are the settings that can be passed to NewServer.
type Options struct {
	BeatmapService cheesegull.BeatmapService
	FileResolver   cheesegull.FileResolver
	Communication  cheesegull.CommunicationService
	APISecret      string
}

// NewServer creates a new HTTP server for CheeseGull.
func NewServer(o Options) http.Handler {
	r := httprouter.New()

	r.GET("/", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
		w.Write([]byte("CheeseGull API " + cheesegull.Version + "\n" +
			"For more information, check out: https://github.com/osuripple/cheesegull"))
	})

	r.GET("/s/:id", o.requestWrapper(old.BeatmapSet))
	r.GET("/b/:id", o.requestWrapper(old.Beatmap))
	r.GET("/index_md5.txt", o.requestWrapper(old.IndexMD5))
	r.GET("/index.json", o.requestWrapper(old.IndexJSON))

	r.POST("/api/request", o.requestWrapperRestr(api.RequestBeatmap))
	r.GET("/api/search", o.requestWrapper(api.Search))
	r.GET("/api/s/:id", o.requestWrapper(old.BeatmapSet))
	r.GET("/api/b/:id", o.requestWrapper(old.Beatmap))

	r.NotFound = _handler{o}

	return r
}

func (o Options) requestWrapper(a func(w http.ResponseWriter, r *http.Request, c *ctx.Context)) func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		o.req(a, w, r, p)
	}
}

// basically the same but checks if APISecret is passed and is valid
func (o Options) requestWrapperRestr(a func(w http.ResponseWriter, r *http.Request, c *ctx.Context)) func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		switch o.APISecret {
		case
			r.URL.Query().Get("k"),
			r.Header.Get("Api-Secret"),
			r.Header.Get("Authorization"):
			o.req(a, w, r, p)
		default:
			w.WriteHeader(403)
			w.Write([]byte("Forbidden"))
		}
	}
}

func (o Options) req(a func(w http.ResponseWriter, r *http.Request, c *ctx.Context), w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	a(w, r, &ctx.Context{
		Params:         p,
		BeatmapService: o.BeatmapService,
		FileResolver:   o.FileResolver,
		Communication:  o.Communication,
		Request:        r,
	})
}
