package ufwb

import (
	"io"
	"os"
)

// BytesFile implements ufwb.File
type BytesFile struct {
	offset int   // Not a int64, since a []byte can only contain `int' elements.
	bytes []byte
}

func NewFileFromBytes (bytes []byte) *BytesFile {
	return &BytesFile{0, bytes}
}

func (f *BytesFile) Read(b []byte) (int, error) {
	n, err := f.ReadAt(b, int64(f.offset))
	f.offset += n
	return n, err
}

func (f *BytesFile) ReadAt(b []byte, off int64) (int, error) {
	if off >= int64(len(f.bytes)) {
		return 0, io.EOF
	}

	n := len(b) // TODO should not be allowed to overun
	copy(b, f.bytes[off:n])

	return n, nil
}

func (f *BytesFile) ReadByte() (byte, error) {
	// TODO No boundary checking!
	b := f.bytes[f.offset]
	f.offset++
	return b, nil
}

func (f *BytesFile) Seek(offset int64, whence int) (ret int64, err error) {
	switch whence {
		case io.SeekStart:
			f.offset = int(offset)
		case io.SeekCurrent:
			f.offset += int(offset)
		case io.SeekEnd:
			f.offset = len(f.bytes) + int(offset)
	}
	if f.offset < 0 {
		f.offset = 0
		return 0, os.ErrInvalid
	}
	return int64(f.offset), nil
}

func (f *BytesFile) Tell() (int64, error) {
	return int64(f.offset), nil
}

