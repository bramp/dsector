package input

import (
	"testing"
)

func TestFileRead(t *testing.T) {
	in, err := OpenOSFile("testdata/8bytes")
	if err != nil {
		t.Fatalf(`OpenOSFile("testdata/8bytes") err = %v want nil`, err)
	}
	testRead(t, in)
}

func TestFileReadShort(t *testing.T) {
	in, err := OpenOSFile("testdata/2bytes")
	if err != nil {
		t.Fatalf(`OpenOSFile("testdata/2bytes") err = %v want nil`, err)
	}
	testReadShort(t, in)
}

func TestFileReadAt(t *testing.T) {
	in, err := OpenOSFile("testdata/8bytes")
	if err != nil {
		t.Fatalf(`OpenOSFile("testdata/8bytes") err = %v want nil`, err)
	}
	testReadAt(t, in)
}

func TestFileSeek(t *testing.T) {
	in, err := OpenOSFile("testdata/8bytes")
	if err != nil {
		t.Fatalf(`OpenOSFile("testdata/8bytes") err = %v want nil`, err)
	}
	testSeek(t, in)
}
