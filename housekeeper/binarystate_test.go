package housekeeper

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var testBeatmaps = []*CachedBeatmap{
	{
		isDownloaded: true,
	},
	{
		ID:           851,
		NoVideo:      true,
		isDownloaded: true,
	},
	{
		ID:           1337777,
		fileSize:     58111,
		isDownloaded: true,
	},
	{
		ID:            851,
		LastUpdate:    time.Date(2017, 9, 21, 11, 11, 50, 0, time.UTC),
		lastRequested: time.Date(2017, 9, 21, 22, 11, 50, 0, time.UTC),
		isDownloaded:  true,
	},
}

func TestEncodeDecode(t *testing.T) {
	buf := &bytes.Buffer{}

	start := time.Now()
	err := writeBeatmaps(buf, testBeatmaps)
	require.NoError(t, err)

	t.Logf("Write took %v", time.Since(start))

	start = time.Now()
	readBMs, err := readBeatmaps(buf)
	require.NoError(t, err)
	t.Logf("Read took %v", time.Since(start))

	require.Equal(t, readBMs, testBeatmaps)
}

func BenchmarkWriteBinaryState(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := writeBeatmaps(fakeWriter{}, testBeatmaps); err != nil {
			panic(err)
		}
	}
}

func BenchmarkReadBinaryState(b *testing.B) {
	buf := &bytes.Buffer{}
	err := writeBeatmaps(buf, testBeatmaps)
	if err != nil {
		b.Fatal(err)
	}
	bufBytes := buf.Bytes()
	bReader := bytes.NewReader(bufBytes)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := readBeatmaps(bReader); err != nil {
			panic(err)
		}
		bReader.Reset(bufBytes)
	}
}

type fakeWriter struct{}

func (fakeWriter) Write(b []byte) (int, error) {
	return len(b), nil
}
