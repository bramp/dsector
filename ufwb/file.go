package ufwb

import (
	"io"
	"os"
)

// OSFile implements ufwb.File by wrapping a os.File
type OSFile os.File

func OpenOSFile(filename string) (*OSFile, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return (*OSFile)(f), nil
}

func NewOSFile(f *os.File) *OSFile {
	return (*OSFile)(f)
}

func (f *OSFile) Tell() (int64, error) {
	return (*os.File)(f).Seek(0, io.SeekCurrent)
}

func (f *OSFile) Read(b []byte) (int, error) {
	return (*os.File)(f).Read(b)
}

func (f *OSFile) ReadAt(b []byte, off int64) (int, error) {
	return (*os.File)(f).ReadAt(b, off)
}

func (f *OSFile) ReadByte() (byte, error) {
	one := make([]byte, 0, 1)
	n, err := f.Read(one)
	if err != nil {
		return '\x00', err
	}
	if n < 1 {
		return '\x00', io.EOF
	}
	return one[0], nil
}

func (f *OSFile) Seek(offset int64, whence int) (ret int64, err error) {
	return (*os.File)(f).Seek(offset, whence)
}

func (f *OSFile) Close() error {
	return (*os.File)(f).Close()
}
