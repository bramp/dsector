package ufwb

import (
	"fmt"
	"encoding/binary"
)


// Value represents one of the parsed elements in the file.
// It doesn't contain the element, just the offset where it starts, and which element it is.
type Value struct {
	Offset  int64 // In bytes from the beginning of the file
	Len     int64 // In bytes
	Element Element

	Children  []*Value
	ByteOrder binary.ByteOrder // Only used for Number, TODO, and TODO. Why have this?
}

func (v *Value) Name() string {
	return v.Element.Name()
}

func (v *Value) Description() string {
	return v.Element.Description()
}

// String returns this value's string representation (based on display, etc)
func (v *Value) Format(f File) (string, error) {
	return v.Element.Format(f, v)
}

func (v *Value) String() string {
	return fmt.Sprintf("[%d len:%d]", v.Offset, v.Len)
}

// Read returns the string representation of this value
//func (v *Value) Read(f File) (string, error) {
//	return "", nil
//}

func (v *Value) Write(f File) {
	panic("TODO")
}

// mustValidiate checks if this Value is valid, and panics if not
func (v *Value) mustValidiate() {
	if err := v.validiate(); err != nil {
		panic(err)
	}
}

// validiate checks if this Value is valid, only used for debugging.
func (v *Value) validiate() error {
	if v == nil {
		return fmt.Errorf("nil Value")
	}
	if v.Len < 0 {
		return fmt.Errorf("Len = %d want >=0", v.Len)
	}
	if v.Offset < 0 {
		return fmt.Errorf("Offset = %d want >0", v.Offset)
	}
	if v.Element == nil {
		return fmt.Errorf("Element = nil want a valid value %v", v)
	}

	end := v.Offset + v.Len
	for _, child := range v.Children {
		if err := child.validiate(); err != nil {
			return err
		}
		if child.Offset < v.Offset || (child.Offset+child.Len) > end {
			return fmt.Errorf("child Value outside of parents bounds: %v > %v", v, child)
		}
	}

	return nil
}
