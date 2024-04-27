package downloader

import (
	"errors"
	"io"
	"strings"
)

var (
	// ErrNoZip is returned from Download when the downloaded file is not a zip file.
	ErrNoZip = errors.New("cheesegull/downloader: file is not a zip archive")

	// ErrNoRedirect is returned from Download when we were not redirect, thus
	// indicating that the beatmap is unavailable.
	ErrNoRedirect = errors.New("cheesegull/downloader: no redirect happened, beatmap could not be downloaded")
)

const zipMagic = "PK\x03\x04"

// Client is an interface that can download a beatmap osz file.
// It returns a ReadCloser that contains the osz file.
type Client interface {
	// HasVideo returns true if the beatmap set has a video.
	// HasVideo(setID int) (bool, error)

	// Download downloads a beatmap set from the remote source.
	Download(setID int, noVideo bool) (io.ReadCloser, error)
}

// Downloader is a wrapper around Client that can download a beatmap.
// It adds a check that the downloaded file is a zip file.
// This should be used rather than DownloaderClient directly.
type Downloader struct {
	Client
}

// NewDownloader returns a new Downloader wrapping the provided DownloaderClient.
func NewDownloader(cl Client) *Downloader {
	return &Downloader{cl}
}

// Download downloads a beatmap set from the remote source using the
// underlying downloaderClient, and checks that the downloaded file is a zip file.
// If the file is not a zip file, errNoZip is returned.
func (d *Downloader) Download(setID int, noVideo bool) (io.ReadCloser, error) {
	body, err := d.Client.Download(setID, noVideo)
	if err != nil {
		return nil, err
	}

	// check that it is a zip file
	first4 := make([]byte, 4)
	_, err = body.Read(first4)
	if err != nil {
		return nil, err
	}
	if string(first4) != zipMagic {
		return nil, ErrNoZip
	}

	// Return zip file and error
	return struct {
		io.Reader
		io.Closer
	}{
		io.MultiReader(strings.NewReader(zipMagic), body),
		body,
	}, nil
}
