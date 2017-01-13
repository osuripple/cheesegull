package cheesegull

import "io"

// BeatmapDownloader is an interface with the task to fetch a beatmap from a
// source.
// First reader is beatmap with video.
// Second reader is beatmap without video.
// If video is not in the beatmap, second reader will be nil and first reader
// will be beatmap without video.
type BeatmapDownloader interface {
	Download(setID int) (io.Reader, io.Reader, error)
}

// FilePlacer is an interface with the task to create an io.Writer in which to
// save the beatmap file. Generally, this would just be a wrapper around
// os.Create, os.Stat, and a few sanity checks to check the destination exists.
type FilePlacer interface {
	Create(n int, noVideo bool) (io.WriteCloser, error)
	Resolve(n int, noVideo bool) string
}
