package fileresolvers

import (
	"io"
	"os"
	"strconv"
)

// FileSystem is a FileResolver that acts on the filesystem.
type FileSystem struct {
	Prefix *string
}

// Create creates a new file in which you can save a beatmap.
func (f FileSystem) Create(n int, noVideo bool) (io.WriteCloser, error) {
	file, err := os.Create(f.Resolve(n, noVideo))
	if os.IsNotExist(err) {
		err = os.MkdirAll(f.prefix(), 0755)
		if err != nil {
			return nil, err
		}
		return os.Create(f.Resolve(n, noVideo))
	}
	return file, err
}

// Open opens a file of the mirror in the filesystem to read its content.
func (f FileSystem) Open(n int, noVideo bool) (io.ReadCloser, error) {
	file, err := os.Open(f.Resolve(n, noVideo))
	if os.IsNotExist(err) {
		return nil, nil
	}
	return file, err
}

// Resolve resolves a file to a name.
func (f FileSystem) Resolve(n int, noVideo bool) string {
	var pf string
	if noVideo {
		pf = "n"
	}
	return f.prefix() + strconv.Itoa(n) + pf + ".osz"
}

const defaultPrefix = "data/"

func (f FileSystem) prefix() string {
	var p string
	if f.Prefix == nil {
		p = defaultPrefix
	} else {
		p = *f.Prefix
	}
	return p
}
