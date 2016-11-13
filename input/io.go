package input

import (
	"io"
)

// ReadUntil reads from the file until it finds a delim byte, EOF, or len bytes have been checked.
// Returning the number of bytes if the delimitor was found, or -1 and a error.
// io.ErrUnexpectedEOF is returned if the EOF was found.
func ReadUntil(r io.ByteReader, delim byte, len int64) (int64, error) {
	var n int64
	for n < len {
		b, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return -1, err
		}
		n++
		if b == delim {
			return n, nil
		}
	}

	return n, io.ErrUnexpectedEOF
}

// ReadAndDiscard reads len bytes from the file, and discards them.
func ReadAndDiscard(r io.Reader, len int64) (int, error) {
	// TODO Change this to use a 4096 byte buffer, that we discard, instead of creating a
	// potentially large slice.
	b := make([]byte, len, len)
	return io.ReadFull(r, b)
}

func ReadFullAt(r io.ReaderAt, buf []byte, off int64) (n int, err error) {
	min := len(buf)
	for n < min && err == nil {
		var nn int
		nn, err = r.ReadAt(buf[n:], off)
		n += nn
		off += int64(nn)
	}
	if n >= min {
		err = nil
	} else if n > 0 && err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return
}
