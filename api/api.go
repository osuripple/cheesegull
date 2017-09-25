// Package api contains the general framework for writing handlers in the
// CheeseGull API.
package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Context is the information that is passed to all request handlers in relation
// to the request, and how to answer it.
type Context struct {
	Request *http.Request
	DB      *sql.DB
	writer  http.ResponseWriter
	params  httprouter.Params
}

// Write writes content to the response body.
func (c *Context) Write(b []byte) (int, error) {
	return c.writer.Write(b)
}

// ReadHeader reads a header from the request.
func (c *Context) ReadHeader(s string) string {
	return c.Request.Header.Get(s)
}

// WriteHeader sets a header in the response.
func (c *Context) WriteHeader(key, value string) {
	c.writer.Header().Set(key, value)
}

// Code sets the response's code.
func (c *Context) Code(i int) {
	c.writer.WriteHeader(i)
}

// Param retrieves a parameter in the URL's path.
func (c *Context) Param(s string) string {
	return c.params.ByName(s)
}

// WriteJSON writes JSON to the response.
func (c *Context) WriteJSON(code int, v interface{}) error {
	c.WriteHeader("Content-Type", "application/json; charset=utf-8")
	c.Code(code)
	return json.NewEncoder(c.writer).Encode(v)
}

type handlerPath struct {
	method, path string
	f            func(c *Context)
}

var handlers []handlerPath

// GET registers a handler for a GET request.
func GET(path string, f func(c *Context)) {
	handlers = append(handlers, handlerPath{"GET", path, f})
}

// POST registers a handler for a POST request.
func POST(path string, f func(c *Context)) {
	handlers = append(handlers, handlerPath{"POST", path, f})
}

// CreateHandler creates a new http.Handler using the handlers registered
// through GET and POST.
func CreateHandler(db *sql.DB) http.Handler {
	r := httprouter.New()
	for _, h := range handlers {
		// Create local copy that we know won't change as the loop proceeds.
		h := h
		r.Handle(h.method, h.path, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			ctx := &Context{
				Request: r,
				DB:      db,
				writer:  w,
				params:  p,
			}
			h.f(ctx)
		})
	}
	return r
}
