package cheesegull

import "io"

// BeatmapDownloader is an interface with the task to fetch a beatmap
// and return an io.Reader that will contain the beatmap data, and an eventual
// error.
type BeatmapDownloader interface {
	Download(setID int) (io.Reader, error)
}
