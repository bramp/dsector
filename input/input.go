// Package input provides various Input implementations for reading from []byte or os.File.
package input

import (
	"fmt"
	"io"
)

type ShortReadError struct {
	n int
}

func (e ShortReadError) Error() string {
	return fmt.Sprintf("short read: %d", e.n)
}

// Input interface provides the minimum methods needed to parse a binary file.
// Thethe interface is changed slightly, by ensuring the Read methods always try to do a full read.
// If they are unable to read the full amount, then a ShortReadError is returned, as well as as much
// as could be read.
type Input interface {
	io.Seeker

	io.Reader
	io.ReaderAt
	io.ByteReader

	Tell() (int64, error) // Here for convenience, perhaps remove.
}
