package input

import (
	"bytes"
	"io"
	"testing"
)

func TestBytesRead(t *testing.T) {
	bytes := []byte{1, 2, 3, 4, 5, 6, 7, 8}

	buf := make([]byte, 1, 1)
	in := FromBytes(bytes)

	for i, want := range bytes {
		n, err := in.Read(buf)
		if n != 1 || err != nil {
			t.Errorf("[%d] in.Read(...) = %d, <%v>, want 1, nil", i, n, err)
			continue
		}
		if buf[0] != want {
			t.Errorf("[%d] buf[0] = %d want %d", i, buf[0], want)
		}
	}

	n, err := in.Read(buf)
	if n != 0 || err != io.EOF {
		t.Errorf("in.Read(...) = %d, <%v>, want 0, io.EOF", n, err)
	}
}

func TestBytesReadShort(t *testing.T) {
	buf := make([]byte, 3, 3)
	in := FromBytes([]byte{1, 2})

	n, err := in.Read(buf)
	if _, ok := err.(ShortReadError); !ok || n != 2 {
		t.Errorf("in.Read(...) = %d, <%v>, want 2, ShortReadError", n, err)
	}
	if bytes.Equal(buf, []byte{1, 2}) {
		t.Errorf("buf = %v want %v", buf[0], []byte{1, 2})
	}

	n, err = in.Read(buf)
	if n != 0 || err != io.EOF {
		t.Errorf("in.Read(...) = %d, <%v>, want 0, io.EOF", n, err)
	}
}

func TestBytesReadAt(t *testing.T) {
	bytes := []byte{1, 2, 3, 4, 5, 6, 7, 8}

	buf := make([]byte, 1, 1)
	in := FromBytes(bytes)

	for _, i := range []int{2, 1, 4, 5, 0, 7, 3, 6} { // Random order
		n, err := in.ReadAt(buf, int64(i))
		if n != 1 || err != nil {
			t.Errorf("[%d] in.Read(...) = %d, <%v>, want 1, nil", i, n, err)
			continue
		}
		if buf[0] != bytes[i] {
			t.Errorf("[%d] buf[0] = %d want %d", i, buf[0], bytes[i])
		}
	}
}
