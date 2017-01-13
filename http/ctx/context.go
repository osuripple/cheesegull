package ctx

import "github.com/julienschmidt/httprouter"
import "github.com/osuripple/cheesegull"

// Context is the information about the request passed to an handler in
// the http package, alongside http.ResponseWriter and *http.Request.
type Context struct {
	Params         httprouter.Params
	BeatmapService cheesegull.BeatmapService
}

// Param retrieves a parameters from the parameters in the Context
func (r *Context) Param(s string) string {
	for _, p := range r.Params {
		if p.Key == s {
			return p.Value
		}
	}
	return ""
}
