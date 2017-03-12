package cheesegull

import (
	"errors"
	"io"
)

// BeatmapDownloader is an interface with the task to fetch a beatmap from a
// source.
// First reader is beatmap with video.
// Second reader is beatmap without video.
// If video is not in the beatmap, second reader will be nil and first reader
// will be beatmap without video.
type BeatmapDownloader interface {
	Download(setID int) (io.ReadCloser, io.ReadCloser, error)
}

// ErrNoRedirect is the error to return when the beatmap, while it is on the
// Source and has a beatmap page, does not actually have a download link, which
// means that the beatmap probably got removed.
var ErrNoRedirect = errors.New("no redirect happened, beatmap could not be downloaded")

// FileResolver is an interface with the task to create an io.Writer in which to
// save the beatmap file. Generally, this would just be a wrapper around
// os.Create, os.Stat, and a few sanity checks to check the destination exists.
type FileResolver interface {
	Create(n int, noVideo bool) (io.WriteCloser, error)
	Open(n int, noVideo bool) (io.ReadCloser, error)
	Resolve(n int, noVideo bool) string
}
