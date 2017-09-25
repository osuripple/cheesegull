package api

func index(c *Context) {
	c.WriteHeader("Content-Type", "text/plain; charset=utf-8")
	c.Write([]byte("CheeseGull v2.x Woo\nFor more information: https://github.com/osuripple/cheesegull"))
}

func init() {
	GET("/", index)
}
