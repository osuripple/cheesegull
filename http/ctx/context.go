package ctx

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"

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
	Logging        cheesegull.Logging
	ErrorHandler   func(error)
	Request        *http.Request
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

// QueryDefault returns a value from the query string. If it's not set, the
// default value provided is used.
func (r *Context) QueryDefault(q string, def string) string {
	v, ok := r.Request.URL.Query()[q]
	if ok && len(v) > 0 {
		return v[0]
	}
	return def
}

// QueryInt gets an integer value from the querystring. Returns 0
// if it's not a valid int.
func (r *Context) QueryInt(q string) int {
	v, _ := strconv.Atoi(r.Request.URL.Query().Get(q))
	return v
}

// QueryIntMultiple is basically like QueryInt, but returns an []int with
// all the ints with the same key passed.
func (r *Context) QueryIntMultiple(q string) []int {
	valsRaw := r.Request.URL.Query()[q]
	vals := make([]int, len(valsRaw))
	for i, v := range valsRaw {
		vals[i], _ = strconv.Atoi(v)
	}
	return vals
}

// QueryIntDefault gets an integer value from the querystring, or if the
// requested parameter is not in the querystring, then it returns a default
// value.
func (r *Context) QueryIntDefault(q string, i int) int {
	v, ok := r.Request.URL.Query()[q]
	if ok && len(v) > 0 {
		ret, _ := strconv.Atoi(v[0])
		return ret
	}
	return i
}
