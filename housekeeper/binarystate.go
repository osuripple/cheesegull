package housekeeper

import (
	"encoding"
	"errors"
	"io"
)

const cachedBeatmapBinSize = 8 + 1 + 15 + 15 + 8

func b2i(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func marshalBinaryCopy(dst []byte, t encoding.BinaryMarshaler) {
	b, _ := t.MarshalBinary()
	copy(dst, b)
}

// copied from binary.BigEndian
func putUint64(b []byte, v uint64) {
	_ = b[7] // boundary check
	b[0] = byte(v >> 56)
	b[1] = byte(v >> 48)
	b[2] = byte(v >> 40)
	b[3] = byte(v >> 32)
	b[4] = byte(v >> 24)
	b[5] = byte(v >> 16)
	b[6] = byte(v >> 8)
	b[7] = byte(v)
}

func writeBeatmaps(w io.Writer, c []*CachedBeatmap) error {
	_, err := w.Write(append([]byte("CGBIN001"), cachedBeatmapBinSize))
	if err != nil {
		return err
	}
	for _, b := range c {
		if b == nil || !b.isDownloaded {
			continue
		}
		enc := make([]byte, cachedBeatmapBinSize)
		putUint64(enc[:8], uint64(b.ID))
		enc[8] = b2i(b.NoVideo)
		marshalBinaryCopy(enc[9:24], b.LastUpdate)
		marshalBinaryCopy(enc[24:39], b.lastRequested)
		putUint64(enc[39:47], b.fileSize)
		_, err := w.Write(enc)
		if err != nil {
			return err
		}
	}
	return nil
}

// copied from binary.BigEndian
func readUint64(b []byte) uint64 {
	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 |
		uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56

}

func readCachedBeatmap(b []byte) *CachedBeatmap {
	m := &CachedBeatmap{}
	m.ID = int(readUint64(b[:8]))
	m.NoVideo = b[8] == 1
	(&m.LastUpdate).UnmarshalBinary(b[9:24])
	(&m.lastRequested).UnmarshalBinary(b[24:39])
	m.fileSize = readUint64(b[39:47])
	m.isDownloaded = true
	return m
}

func readBeatmaps(r io.Reader) ([]*CachedBeatmap, error) {
	b := make([]byte, 8)
	_, err := r.Read(b)
	if err != nil {
		return nil, err
	}
	if string(b) != "CGBIN001" {
		return nil, errors.New("cheesegull/housekeeper: unknown cgbin version")
	}

	b = make([]byte, 1)
	_, err = r.Read(b)
	if err != nil {
		return nil, err
	}

	bmLength := b[0]
	if bmLength == 0 {
		return nil, nil
	}
	b = make([]byte, bmLength)
	beatmaps := make([]*CachedBeatmap, 0, 50)

	for {
		read, err := r.Read(b)
		switch {
		case err == io.EOF:
			return beatmaps, nil
		case err != nil:
			return nil, err
		case byte(read) != bmLength:
			return beatmaps, nil
		}
		beatmaps = append(beatmaps, readCachedBeatmap(b))
	}
}
