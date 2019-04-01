package housekeeper

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

// CachedBeatmap represents a beatmap that is held in the cache of CheeseGull.
type CachedBeatmap struct {
	ID         int
	NoVideo    bool
	LastUpdate time.Time

	lastRequested time.Time

	fileSize     uint64
	isDownloaded bool
	mtx          sync.RWMutex
	waitGroup    sync.WaitGroup
}

// File opens the File of the beatmap from the filesystem.
func (c *CachedBeatmap) File() (*os.File, error) {
	return os.Open(c.fileName())
}

// CreateFile creates the File of the beatmap in the filesystem, and returns it
// in write mode.
func (c *CachedBeatmap) CreateFile() (*os.File, error) {
	return os.Create(c.fileName())
}

func (c *CachedBeatmap) fileName() string {
	n := ""
	if c.NoVideo {
		n = "n"
	}
	return "data/" + strconv.Itoa(c.ID) + n + ".osz"
}

// IsDownloaded checks whether the beatmap has been downloaded.
func (c *CachedBeatmap) IsDownloaded() bool {
	c.mtx.RLock()
	i := c.isDownloaded
	c.mtx.RUnlock()
	return i
}

// FileSize returns the FileSize of c.
func (c *CachedBeatmap) FileSize() uint64 {
	c.mtx.RLock()
	i := c.fileSize
	c.mtx.RUnlock()
	return i
}

// MustBeDownloaded will check whether the beatmap is downloaded.
// If it is not, it will wait for it to become downloaded.
func (c *CachedBeatmap) MustBeDownloaded() {
	if c.IsDownloaded() {
		return
	}
	c.waitGroup.Wait()
}

// DownloadCompleted must be called once the beatmap has finished downloading.
func (c *CachedBeatmap) DownloadCompleted(fileSize uint64, parentHouse *House) {
	c.mtx.Lock()
	c.fileSize = fileSize
	c.isDownloaded = true
	c.mtx.Unlock()
	c.waitGroup.Done()
	parentHouse.scheduleCleanup()
}

// SetLastRequested changes the last requested time.
func (c *CachedBeatmap) SetLastRequested(t time.Time) {
	c.mtx.Lock()
	c.lastRequested = t
	c.mtx.Unlock()
}

func (c *CachedBeatmap) String() string {
	return fmt.Sprintf("{ID: %d NoVideo: %t LastUpdate: %v}", c.ID, c.NoVideo, c.LastUpdate)
}

// AcquireBeatmap attempts to add a new CachedBeatmap to the state.
// In order to add a new CachedBeatmap to the state, one must not already exist
// in the state with the same ID, NoVideo and LastUpdate. In case one is already
// found, this is returned, alongside with false. If LastUpdate is newer than
// that of the beatmap stored in the state, then the beatmap in the state's
// downloaded status is switched back to false and the LastUpdate is changed.
// true is also returned, indicating that the caller now has the burden of
// downloading the beatmap.
//
// In the case the cachedbeatmap has not been stored in the state, then
// it is added to the state and, like the case where LastUpdated has been
// changed, true is returned, indicating that the caller must now download the
// beatmap.
//
// If you're confused attempting to read this, let me give you an example:
//
//   A: Yo, is this beatmap cached?
//   B: Yes, yes it is! Here you go with the information about it. No need to do
//      anything else.
//      ----
//   A: Yo, got this beatmap updated 2 hours ago. Have you got it cached?
//   B: Ah, I'm afraid that I only have the version updated 10 hours ago.
//      Mind downloading the updated version for me?
//      ----
//   A: Yo, is this beatmap cached?
//   B: Nope, I didn't know it existed before you told me. I've recorded its
//      info now, but jokes on you, you now have to actually download it.
//      Chop chop!
func (h *House) AcquireBeatmap(c *CachedBeatmap) (*CachedBeatmap, bool) {
	if c == nil {
		return nil, false
	}

	h.stateMutex.Lock()
	for _, b := range h.state {
		// if the id or novideo is different, then all is good and we
		// can proceed with the next element.
		if b.ID != c.ID || b.NoVideo != c.NoVideo {
			continue
		}
		// unlocking because in either branch, we will return.
		h.stateMutex.Unlock()

		b.mtx.Lock()
		// if c is not newer than b, then just return.
		if !b.LastUpdate.Before(c.LastUpdate) {
			b.mtx.Unlock()
			return b, false
		}

		b.LastUpdate = c.LastUpdate
		b.mtx.Unlock()
		b.waitGroup.Add(1)
		return b, true
	}

	// c was not present in our state: we need to add it.

	// we need to recreate the CachedBeatmap: this way we can be sure the zero
	// is set for the unexported fields.
	n := &CachedBeatmap{
		ID:         c.ID,
		NoVideo:    c.NoVideo,
		LastUpdate: c.LastUpdate,
	}
	h.state = append(h.state, n)
	h.stateMutex.Unlock()

	n.waitGroup.Add(1)
	return n, true
}
