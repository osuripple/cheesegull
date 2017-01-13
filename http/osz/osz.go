// Package osz handles retrieval of osz files.
package osz

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"os"

	"github.com/osuripple/cheesegull/http/ctx"
)

// GetBeatmap allows to retrieve a beatmap from the mirror.
func GetBeatmap(w http.ResponseWriter, r *http.Request, c *ctx.Context) {
	// Get the necessary part, first of all.
	f := strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, ".osz"), "/")
	if f == "" {
		return
	}

	// will be used later with Content-Disposition
	originalF := f

	// Detect if we got a request for novideo.
	noVideo := f[len(f)-1] == 'n'
	if noVideo {
		f = f[:len(f)-1]
	}

	// Get the ID
	id, _ := strconv.Atoi(f)

	// Open the file
	file, err := c.FileResolver.Open(id, noVideo)
	if err != nil {
		c.HandleError(err)
		return
	}

	// If it did not work with "n", it might mean the beatmap has got no video
	// at all, so let's try without.
	if file == nil && noVideo {
		file, err = c.FileResolver.Open(id, false)
		if err != nil {
			c.HandleError(err)
			return
		}
	}

	if file == nil {
		w.WriteHeader(404)
		w.Write([]byte("That beatmap could not be found :("))
		return
	}

	defer file.Close()

	// All checks passed through. We are now ready to transmit our beatmap.

	// Transmit content-length if we can
	if fix, ok := file.(interface {
		Stat() (os.FileInfo, error)
	}); ok {
		finfo, _ := fix.Stat()
		w.Header().Add("Content-Length", fmt.Sprintf("%d", finfo.Size()))
	}
	w.Header().Add("Content-Type", "application/octet-stream")
	w.Header().Add("Content-Disposition", fmt.Sprintf(`attachment;filename="%s.osz"`, originalF))

	_, err = io.Copy(w, file)
	if err != nil {
		c.HandleError(err)
	}
}
