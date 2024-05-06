package downloader

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"strings"
	"time"
)

var (
	// ErrNoZip is returned from Download when the downloaded file is not a zip file.
	ErrNoZip = errors.New("cheesegull/downloader: file is not a zip archive")

	// ErrNoRedirect is returned from Download when we were not redirect, thus
	// indicating that the beatmap is unavailable.
	ErrNoRedirect = errors.New("cheesegull/downloader: no redirect happened, beatmap could not be downloaded")

	// ErrTemporaryFailure is returned from Download when the download failed, like when it returns a 503.
	ErrTemporaryFailure = errors.New("temporary failure")
)

const (
	zipMagic   = "PK\x03\x04"
	maxRetries = 5
	retryDelay = time.Millisecond * 500
)

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

func (d *Downloader) delayForRetry(retries int) time.Duration {
	if retries < 0 {
		retries = 0
	}
	if retries > maxRetries {
		retries = maxRetries
	}
	return time.Duration(math.Pow(2, float64(retries))) * retryDelay
}

// Download downloads a beatmap set from the remote source using the
// underlying downloaderClient, and checks that the downloaded file is a zip file.
// If the file is not a zip file, errNoZip is returned.
// If the underlying downloader returns an error wrapping ErrTemporaryFailure, the request is
// retries up to maxRetries times with an exponential backoff.
func (d *Downloader) Download(setID int, noVideo bool) (io.ReadCloser, error) {
	var body io.ReadCloser
	var err error
	var retries int
	var ok bool

	var downstreamErr error
	for !ok {
		body, err = d.Client.Download(setID, noVideo)
		if errors.Is(err, ErrTemporaryFailure) {
			downstreamErr = errors.Join(downstreamErr, err)
			if retries >= maxRetries {
				return nil, fmt.Errorf("too many temporary failures, giving up. original error: %w", downstreamErr)
			}
			delay := d.delayForRetry(retries)
			log.Printf("Temporary failure (%q), retrying in %v", err, delay)
			time.Sleep(delay)
			retries += 1
			continue
		}

		if err != nil {
			return nil, err
		}

		ok = true
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
