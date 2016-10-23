// Package input provides various Input implementations for reading from []byte or os.File.
package input

import "io"

// Input interface provides the minimum methods needed to parse a binary file.
type Input interface {
	io.Seeker

	io.Reader
	io.ReaderAt
	io.ByteReader

	Tell() (int64, error) // Here for convenience, perhaps remove.
}
