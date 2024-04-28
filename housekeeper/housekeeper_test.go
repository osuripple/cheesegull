package housekeeper

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testFile struct {
	path string
}

func newTestFolder(t *testing.T) testFile {
	return testFile{
		path: filepath.Join(t.TempDir(), "cgbin.db"),
	}
}

func (f testFile) cleanup(t *testing.T) func() {
	return func() {
		err := os.Remove(f.path)
		require.NoError(t, err)
	}
}

func TestCleanup(t *testing.T) {
	expectRemove := []*CachedBeatmap{
		{
			ID:            1,
			lastRequested: time.Date(2017, 4, 5, 15, 5, 3, 0, time.UTC),
			fileSize:      15000,
			isDownloaded:  true,
		},
	}
	expectRemain := []*CachedBeatmap{
		{
			ID:            2,
			lastRequested: time.Date(2017, 4, 10, 15, 5, 3, 0, time.UTC),
			fileSize:      15000,
			isDownloaded:  true,
		},
		{
			ID:            3,
			lastRequested: time.Date(2017, 4, 15, 15, 5, 3, 0, time.UTC),
			fileSize:      15000,
			isDownloaded:  true,
		},
		{
			ID:            4,
			lastRequested: time.Date(2017, 4, 20, 15, 5, 3, 0, time.UTC),
			fileSize:      15000,
			isDownloaded:  true,
		},
	}

	f := newTestFolder(t)
	t.Cleanup(f.cleanup(t))

	h := New(f.path)
	h.MaxSize = 50000
	h.state = append(expectRemain, expectRemove...)
	h.dryRun = make([]*CachedBeatmap, 0)

	start := time.Now()
	h.cleanUp()
	t.Log("cleanup took", time.Since(start))

	require.Equal(t, expectRemain, h.state)
	require.Equal(t, expectRemove, h.dryRun)
}

func TestCleanupNoMaps(t *testing.T) {
	expectRemove := []*CachedBeatmap{}
	expectRemain := []*CachedBeatmap{
		{
			ID:            1,
			lastRequested: time.Date(2017, 4, 10, 15, 5, 3, 0, time.UTC),
			fileSize:      100000,
			isDownloaded:  true,
		},
	}

	f := newTestFolder(t)
	t.Cleanup(f.cleanup(t))

	h := New(f.path)
	h.MaxSize = 100000
	h.state = append(expectRemain, expectRemove...)
	h.dryRun = make([]*CachedBeatmap, 0)

	start := time.Now()
	h.cleanUp()
	t.Log("cleanup took", time.Since(start))

	require.Equal(t, expectRemain, h.state)
	require.Empty(t, h.dryRun)
}

func TestCleanupEmptyBeatmaps(t *testing.T) {
	expectRemove := []*CachedBeatmap{
		{
			ID:            1,
			lastRequested: time.Date(2017, 4, 10, 15, 5, 3, 0, time.UTC),
			fileSize:      10,
			isDownloaded:  true,
		},
	}
	expectRemain := []*CachedBeatmap{
		{
			ID:            2,
			lastRequested: time.Date(2017, 4, 5, 15, 5, 3, 0, time.UTC),
			fileSize:      0,
			isDownloaded:  true,
		},
		{
			ID:            3,
			lastRequested: time.Date(2017, 4, 4, 15, 5, 3, 0, time.UTC),
			fileSize:      0,
			isDownloaded:  true,
		},
		{
			ID:            4,
			lastRequested: time.Date(2017, 4, 3, 15, 5, 3, 0, time.UTC),
			fileSize:      0,
			isDownloaded:  true,
		},
	}

	f := newTestFolder(t)
	t.Cleanup(f.cleanup(t))

	h := New(f.path)
	h.MaxSize = 5
	h.state = append(expectRemain, expectRemove...)
	h.dryRun = make([]*CachedBeatmap, 0)

	start := time.Now()
	h.cleanUp()
	t.Log("cleanup took", time.Since(start))

	require.Equal(t, expectRemain, h.state)
	require.Equal(t, expectRemove, h.dryRun)
}
