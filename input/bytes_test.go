package input

import (
	"testing"
)

func TestBytesRead(t *testing.T) {
	in := FromBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8})
	testRead(t, in)
}

func TestBytesReadShort(t *testing.T) {
	in := FromBytes([]byte{1, 2})
	testReadShort(t, in)
}

func TestBytesReadAt(t *testing.T) {
	in := FromBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8})
	testReadAt(t, in)
}

func TestBytesSeek(t *testing.T) {
	in := FromBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8})
	testSeek(t, in)
}
