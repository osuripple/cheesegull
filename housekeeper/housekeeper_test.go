package housekeeper

import (
	"reflect"
	"testing"
	"time"
)

func TestCleanupOneMap(t *testing.T) {
	expectRemove := []*CachedBeatmap{
		&CachedBeatmap{
			ID:            1,
			lastRequested: time.Date(2017, 4, 5, 15, 5, 3, 0, time.UTC),
			fileSize:      15,
			isDownloaded:  true,
		},
	}
	expectRemain := []*CachedBeatmap{
		&CachedBeatmap{
			ID:            2,
			lastRequested: time.Date(2017, 4, 10, 15, 5, 3, 0, time.UTC),
			fileSize:      15,
			isDownloaded:  true,
		},
		&CachedBeatmap{
			ID:            3,
			lastRequested: time.Date(2017, 4, 15, 15, 5, 3, 0, time.UTC),
			fileSize:      15,
			isDownloaded:  true,
		},
		&CachedBeatmap{
			ID:            4,
			lastRequested: time.Date(2017, 4, 20, 15, 5, 3, 0, time.UTC),
			fileSize:      15,
			isDownloaded:  true,
		},
	}

	h := New()
	h.MaxSize = 50
	h.state = append(expectRemain, expectRemove...)
	h.dryRun = make([]*CachedBeatmap, 0)

	start := time.Now()
	h.cleanUp()
	t.Log("cleanup took", time.Since(start))

	if !reflect.DeepEqual(expectRemain, h.state) {
		t.Errorf("Want %v got %v", expectRemain, h.state)
	}
	if !reflect.DeepEqual(expectRemove, h.dryRun) {
		t.Errorf("Want %v got %v", expectRemove, h.dryRun)
	}
}

func TestCleanupNoMaps(t *testing.T) {
	expectRemove := []*CachedBeatmap{}
	expectRemain := []*CachedBeatmap{
		&CachedBeatmap{
			ID:            1,
			lastRequested: time.Date(2017, 4, 10, 15, 5, 3, 0, time.UTC),
			fileSize:      10,
			isDownloaded:  true,
		},
	}

	h := New()
	h.MaxSize = 10
	h.state = append(expectRemain, expectRemove...)
	h.dryRun = make([]*CachedBeatmap, 0)

	start := time.Now()
	h.cleanUp()
	t.Log("cleanup took", time.Since(start))

	if !reflect.DeepEqual(expectRemain, h.state) {
		t.Errorf("Want %v got %v", expectRemain, h.state)
	}
	if !reflect.DeepEqual(expectRemove, h.dryRun) {
		t.Errorf("Want %v got %v", expectRemove, h.dryRun)
	}
}

func TestCleanupEmptyBeatmaps(t *testing.T) {
	expectRemove := []*CachedBeatmap{
		&CachedBeatmap{
			ID:            1,
			lastRequested: time.Date(2017, 4, 10, 15, 5, 3, 0, time.UTC),
			fileSize:      10,
			isDownloaded:  true,
		},
	}
	expectRemain := []*CachedBeatmap{
		&CachedBeatmap{
			ID:            2,
			lastRequested: time.Date(2017, 4, 5, 15, 5, 3, 0, time.UTC),
			fileSize:      0,
			isDownloaded:  true,
		},
		&CachedBeatmap{
			ID:            3,
			lastRequested: time.Date(2017, 4, 4, 15, 5, 3, 0, time.UTC),
			fileSize:      0,
			isDownloaded:  true,
		},
		&CachedBeatmap{
			ID:            4,
			lastRequested: time.Date(2017, 4, 3, 15, 5, 3, 0, time.UTC),
			fileSize:      0,
			isDownloaded:  true,
		},
	}

	h := New()
	h.MaxSize = 5
	h.state = append(expectRemain, expectRemove...)
	h.dryRun = make([]*CachedBeatmap, 0)

	start := time.Now()
	h.cleanUp()
	t.Log("cleanup took", time.Since(start))

	if !reflect.DeepEqual(expectRemain, h.state) {
		t.Errorf("Want %v got %v", expectRemain, h.state)
	}
	if !reflect.DeepEqual(expectRemove, h.dryRun) {
		t.Errorf("Want %v got %v", expectRemove, h.dryRun)
	}
}
