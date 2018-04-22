package api

import (
	"expvar"
)

// Version is set by main and it is given to requests at /
var Version = "v2.DEV"

func index(c *Context) {
	c.WriteHeader("Content-Type", "text/plain; charset=utf-8")
	c.Write([]byte("CheeseGull " + Version + " Woo\nFor more information: https://github.com/osuripple/cheesegull"))
}

var _evh = expvar.Handler()

func expvarHandler(c *Context) {
	_evh.ServeHTTP(c.writer, c.Request)
}

func init() {
	GET("/", index)
	GET("/expvar", expvarHandler)
}
