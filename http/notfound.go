package http

import (
	"net/http"
	"regexp"

	"github.com/osuripple/cheesegull/http/ctx"
	"github.com/osuripple/cheesegull/http/osz"
)

type _handler struct {
	o Options
}

func (h _handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	notFoundHandler(w, r, &ctx.Context{
		BeatmapService: h.o.BeatmapService,
		FileResolver:   h.o.FileResolver,
	})
}

// Yes, I should probably use some better validation for this but can't be
// bothered.
var beatmapRegex = regexp.MustCompile(`^/[0-9]+n?\.osz$`)

func notFoundHandler(w http.ResponseWriter, r *http.Request, c *ctx.Context) {
	if beatmapRegex.MatchString(r.URL.Path) {
		osz.GetBeatmap(w, r, c)
		return
	}
	w.WriteHeader(404)
	w.Write([]byte("404 Not Found"))
}
