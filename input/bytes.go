// TODO Write tests
package input

import (
	"io"
	"os"
)

// Bytes implements Input
type Bytes struct {
	offset int // Not a int64, since a []byte can only contain `int' elements.
	bytes  []byte
}

func FromBytes(bytes []byte) *Bytes {
	return &Bytes{0, bytes}
}

func (f *Bytes) Read(b []byte) (int, error) {
	n, err := f.ReadAt(b, int64(f.offset))
	f.offset += n
	return n, err
}

func (f *Bytes) ReadAt(b []byte, off int64) (int, error) {
	if off < 0 || off >= int64(len(f.bytes)) {
		return 0, io.EOF
	}

	end := int(off) + len(b)
	if end > len(f.bytes) {
		end = len(f.bytes)
		return (end - int(off)), io.ErrUnexpectedEOF
	}

	return copy(b, f.bytes[off:end]), nil
}

func (f *Bytes) ReadByte() (byte, error) {
	if f.offset < 0 || f.offset >= len(f.bytes) {
		return 0, io.EOF
	}

	b := f.bytes[f.offset]
	f.offset++
	return b, nil
}

func (f *Bytes) Seek(offset int64, whence int) (int64, error) {
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

func (f *Bytes) Tell() (int64, error) {
	return int64(f.offset), nil
}
