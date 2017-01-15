package ctx

import (
	"fmt"
	"runtime/debug"

	"github.com/julienschmidt/httprouter"
	"github.com/osuripple/cheesegull"
)

// Context is the information about the request passed to an handler in
// the http package, alongside http.ResponseWriter and *http.Request.
type Context struct {
	Params         httprouter.Params
	BeatmapService cheesegull.BeatmapService
	FileResolver   cheesegull.FileResolver
	Communication  cheesegull.CommunicationService
	ErrorHandler   func(error)
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

// HandleError handles an error, passing it to ErrorHandler or printing it and
// printing the stacktrace.
func (r *Context) HandleError(err error) {
	if r.ErrorHandler != nil {
		r.ErrorHandler(err)
		return
	}
	fmt.Println(err)
	debug.PrintStack()
}
