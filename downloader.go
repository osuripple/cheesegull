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
