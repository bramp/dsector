package input

import (
	"bytes"
	"io"
	"testing"
)

func testReadShort(t *testing.T, in Input) {
	// Try to read 3 bytes, from a 2 byte array
	buf := make([]byte, 3, 3)
	n, err := in.Read(buf)
	if err != io.ErrUnexpectedEOF || n != 2 {
		t.Errorf("in.Read(...) = %d, <%v>, want 2, <ErrUnexpectedEOF>", n, err)
	}
	if bytes.Equal(buf, []byte{1, 2}) {
		t.Errorf("buf = %v want %v", buf[0], []byte{1, 2})
	}

	// Try and read again (should be at EOF)
	n, err = in.Read(buf)
	if n != 0 || err != io.EOF {
		t.Errorf("in.Read(...) = %d, <%v>, want 0, <EOF>", n, err)
	}
}

func testRead(t *testing.T, in Input) {
	// Read one at a time
	buf := make([]byte, 1, 1)
	for i, want := range []byte{1, 2, 3, 4, 5, 6, 7, 8} {
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

func testReadAt(t *testing.T, in Input) {
	buf := make([]byte, 1, 1)

	for _, i := range []int{2, 1, 4, 5, 0, 7, 3, 6} { // Random order
		n, err := in.ReadAt(buf, int64(i))
		if n != 1 || err != nil {
			t.Errorf("[%d] in.Read(...) = %d, <%v>, want 1, nil", i, n, err)
			continue
		}
		if buf[0] != (byte)(i+1) {
			t.Errorf("[%d] buf[0] = %d want %d", i, buf[0], (i + 1))
		}
	}
}

func testSeek(t *testing.T, in Input) {
	n, err := in.Tell()
	if n != 0 || err != nil {
		t.Errorf("in.Tell(...) = %d, <%v>, want 0, nil", n, err)
	}

	n, err = in.Seek(0, io.SeekStart)
	if n != 0 || err != nil {
		t.Errorf("in.Seek(0, io.SeekStart) = %d, <%v>, want 0, nil", n, err)
	}

	n, err = in.Tell()
	if n != 0 || err != nil {
		t.Errorf("in.Tell(...) = %d, <%v>, want 0, nil", n, err)
	}

	n, err = in.Seek(4, io.SeekStart)
	if n != 4 || err != nil {
		t.Errorf("in.Seek(4, io.SeekStart) = %d, <%v>, want 4, nil", n, err)
	}

	n, err = in.Tell()
	if n != 4 || err != nil {
		t.Errorf("in.Tell(...) = %d, <%v>, want 4, nil", n, err)
	}

	n, err = in.Seek(4, io.SeekCurrent)
	if n != 8 || err != nil {
		t.Errorf("in.Seek(4, io.SeekCurrent) = %d, <%v>, want 8, nil", n, err)
	}

	n, err = in.Tell()
	if n != 8 || err != nil {
		t.Errorf("in.Tell(...) = %d, <%v>, want 8, nil", n, err)
	}

	// Go past the end of the file
	n, err = in.Seek(1, io.SeekCurrent)
	if n != 9 || err != nil {
		t.Errorf("in.Seek(1, io.SeekCurrent) = %d, <%v>, want 9, nil", n, err)
	}

	n, err = in.Tell()
	if n != 9 || err != nil {
		t.Errorf("in.Tell(...) = %d, <%v>, want 9, nil", n, err)
	}

	// Reset
	n, err = in.Seek(0, io.SeekStart)
	if n != 0 || err != nil {
		t.Errorf("in.Seek(0, io.SeekStart) = %d, <%v>, want 0, nil", n, err)
	}

	n, err = in.Tell()
	if n != 0 || err != nil {
		t.Errorf("in.Tell(...) = %d, <%v>, want 0, nil", n, err)
	}

	n, err = in.Seek(0, io.SeekEnd)
	if n != 8 || err != nil {
		t.Errorf("in.Seek(0, io.SeekEnd) = %d, <%v>, want 8, nil", n, err)
	}

	n, err = in.Tell()
	if n != 8 || err != nil {
		t.Errorf("in.Tell(...) = %d, <%v>, want 8, nil", n, err)
	}
}
