package downloader

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDelayForRetry(t *testing.T) {
	d := NewDownloader(nil)
	for _, tc := range []struct {
		retries     int
		expDuration time.Duration
	}{
		{0, time.Millisecond * 500},
		{1, time.Second * 1},
		{2, time.Second * 2},
		{3, time.Second * 4},
		{4, time.Second * 8},
		{5, time.Second * 16},

		{-1, time.Millisecond * 500},
		{500, time.Second * 16},
	} {
		t.Run(strconv.Itoa(tc.retries), func(t *testing.T) {
			assert.Equal(t, tc.expDuration, d.delayForRetry(tc.retries))
		})
	}
}
