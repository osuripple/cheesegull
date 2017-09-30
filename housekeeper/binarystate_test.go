package housekeeper

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

var testBeatmaps = []*CachedBeatmap{
	&CachedBeatmap{
		isDownloaded: true,
	},
	&CachedBeatmap{
		ID:           851,
		NoVideo:      true,
		isDownloaded: true,
	},
	&CachedBeatmap{
		ID:           1337777,
		fileSize:     58111,
		isDownloaded: true,
	},
	&CachedBeatmap{
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
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Write took %v", time.Since(start))

	start = time.Now()
	readBMs, err := readBeatmaps(buf)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Read took %v", time.Since(start))

	if !reflect.DeepEqual(readBMs, testBeatmaps) {
		t.Fatalf("original %v read %v", testBeatmaps, readBMs)
	}
}

func BenchmarkWriteBinaryState(b *testing.B) {
	for i := 0; i < b.N; i++ {
		writeBeatmaps(fakeWriter{}, testBeatmaps)
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
		readBeatmaps(bReader)
		bReader.Reset(bufBytes)
	}
}

type fakeWriter struct{}

func (fakeWriter) Write(b []byte) (int, error) {
	return len(b), nil
}
