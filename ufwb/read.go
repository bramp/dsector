package ufwb

// This file uses the grammar to parse the binary file.

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"bytes"
)

type Decoder struct {
	u *Ufwb
	f File

	// dynamicEndian be changed by scripts during processing.
	dynamicEndian binary.ByteOrder
}

func NewDecoder(u *Ufwb, f File) *Decoder {
	return &Decoder{
		u: u,
		f: f,
	}
}

func (d *Decoder) ByteOrder(e Endian) binary.ByteOrder {
	if e == LittleEndian {
		return binary.LittleEndian
	} else if e == BigEndian {
		return binary.BigEndian
	} else if e == DynamicEndian {
		return d.dynamicEndian
	}
	panic(fmt.Sprintf("Unknown endian %v", e))
}

func (d *Decoder) String() string {
	pos, err := d.f.Tell();
	if err != nil {
		return fmt.Sprintf("%s", err)
	}
	return fmt.Sprintf("%d", pos)
}

// File interface provides the minimum needed to parse the binary file.
type File interface {
	io.Seeker

	io.Reader
	io.ReaderAt
	io.ByteReader

	Tell() (int64, error) // Here for convinence, perhaps remove.
}

func (d *Decoder) Decode() (*Value, error) {
	return d.u.Read(d)
}

// readAssert wraps Read() calls to ensure they are valid (to catch programmer mistakes)
func (d *Decoder) readAssert(v *Value, e error) (*Value, error) {
	if v != nil {
		v.mustValidiate()
		pos, err := d.f.Tell()
		if err != nil {
			if (v.Offset + v.Len) != pos {
				panic(fmt.Sprintf("Decoder was not left at right position after %v", v))
			}
		}
	}
	return v, e
}

func (u *Ufwb) Read(d *Decoder) (*Value, error) {
	log.Debugf("[%s] Ufwb Read", d)
	return d.readAssert(u.Grammar.Read(d))
}

func (g *Grammar) Read(d *Decoder) (*Value, error) {
	log.Debugf("[%s] Read Grammar%s", d, g.Base.String())
	return d.readAssert(g.Start.Read(d))
}

func (s *Structure) Read(d *Decoder) (*Value, error) {
	log.Debugf("[%s] Read Structure%s", d, s.Base.String())

	start, err := d.f.Tell()
	if err != nil {
		return nil, err
	}

	if start > 10000 {
		panic("DEBUG: ENDING EARLY")
	}

	length := int64(0)
	var children []*Value

	elements := s.Elements()
	for i := 0; i < len(elements); i++ {
		startElement, err := d.f.Tell()
		if err != nil {
			return nil, err
		}

		e := elements[i]

		v, err := d.readAssert(e.Read(d))
		if err == io.EOF {
			break
		}

		if err != nil {
			switch s.Order() {
			case FixedOrder:
				return nil, err
			case VariableOrder:
				log.Debugf("%v failed %s", e, err)

				// Reset and try next element
				d.f.Seek(startElement, io.SeekStart)
				continue
			default:
				return nil, &validationError{e: s, msg: fmt.Sprintf("unknown order: %s", s.Order())}
			}
		}
		length += v.Len
		children = append(children, v)

		if s.Order() == VariableOrder {
			// If we are variable order, start again from the first element for the next round
			i = 0
			// TODO Detect if we are in parsing loop
		}
	}

	// TODO Check if we reached the min children

	return &Value{
		Offset:   start,
		Len:      length,
		Element:  s,
		Children: children,
	}, nil
}

// TODO Make this actually eval the string, and determine the right value
func (d *Decoder) eval(s Reference) int64 {
	i, err := strconv.Atoi(string(s))
	if err != nil {
		panic(err) // TODO REMOVE PANIC
	}
	return int64(i)
}

func (s *String) Read(d *Decoder) (*Value, error) {
	log.Debugf("[%s] Read String%s", d, s.Base.String())

	start, err := d.f.Tell()
	if err != nil {
		return nil, err
	}

	switch s.Typ() {
	case "zero-terminated":
		n, err := seekUntil(d.f, '\x00')
		if err != nil {
			return nil, err
		}
		return &Value{Offset: start, Len: n, Element: s}, nil

	case "fixed-length":
		len := d.eval(s.Length())
		d.f.Seek(len, io.SeekCurrent)
		return &Value{Offset: start, Len: len, Element: s}, nil
	}

	return nil, fmt.Errorf("unknown type %q", s.Typ())
}

// skip moves the decoder forward by length, returning a Value that covers the range
func skip(d *Decoder, length Reference, lengthUnit LengthUnit) (*Value, error) {
	start, err := d.f.Tell()
	if err != nil {
		return nil, err
	}

	len := d.eval(length)

	if lengthUnit == BitLengthUnit {
		len /= 8 // TODO Test edge cases (such as 20 bits)
	} else if lengthUnit == ByteLengthUnit {
		// Do nothing
	} else {
		return nil, fmt.Errorf("unknown length unit: %s", lengthUnit)
	}

	_, err = d.f.Seek(len, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	return &Value{
		Offset: start,
		Len:    len,
	}, nil
}

func (b *Binary) Read(d *Decoder) (*Value, error) {
	log.Debugf("[%s] Read Binary%s", d, b.Base.String())

	v, err := skip(d, b.Length(), b.lengthUnit)
	if err != nil {
		return nil, err
	}

	v.Element = b

	// If we have FixedValues, then check at least one matches
	values := b.Values()
	if len(values) > 0 { // TODO Perhaps change with n.MustMatch()
		// Read the bytes value
		bs, err := b.Bytes(d.f, v)
		if err != nil {
			return nil, err
		}
		// Now check it matches one of the fixed values
		for _, fv := range values {
			if bytes.Equal(fv.value, bs) {
				return v, nil
			}
		}
		f, _ := b.format(bs)
		return v, &validationError{e: b, msg: fmt.Sprintf("%q does match any of the fixed values", f)}
	}


	return v, nil
}

func (b *Binary) Bytes(f File, value *Value) ([]byte, error) {
	out := make([]byte, value.Len, value.Len)
	n, err := f.ReadAt(out, value.Offset)
	return out[:n], err
}

// int returns the integer stored at Value in f. The returned
// integer is one of int{8,16,32,64} or uint{8,16,32,64} depending
// on the width and sign of the integer.
func (n *Number) int(f File, value *Value) (interface{}, error) {
	if _, err := f.Seek(value.Offset, io.SeekStart); err != nil {
		return 0, err
	}

	var i interface{}
	if n.signed {
		switch value.Len {
		case 1:
			i = new(int8)
		case 2:
			i = new(int16)
		case 4:
			i = new(int32)
		case 8:
			i = new(int64)
		}
	} else {
		switch value.Len {
		case 1:
			i = new(uint8)
		case 2:
			i = new(uint16)
		case 4:
			i = new(uint32)
		case 8:
			i = new(uint64)
		}
	}
	if i == nil {
		return 0, fmt.Errorf("unsupported number length: %d", value.Len)
	}

	err := binary.Read(f, value.ByteOrder, i)

	// Strip the pointer from the interface
	// TODO Consider refactoring, so ints are always either int64 or uint64
	return reflect.ValueOf(i).Elem().Interface() , err
}

// Int returns the int this file/value refers to.
func (n *Number) Int(f File, value *Value) (int64, error) {
	i, err := n.int(f, value)
	if err != nil {
		return 0, err
	}
	return reflect.ValueOf(i).Int(), nil
}

func (n *Number) Uint(f File, value *Value) (uint64, error) {
	i, err := n.int(f, value)
	if err != nil {
		return 0, err
	}
	return reflect.ValueOf(i).Uint(), nil
}

func (n *Number) Read(d *Decoder) (*Value, error) {
	log.Debugf("[%s] Read Number%s", d, n.Base.String())

	v, err := skip(d, n.Length(), n.LengthUnit())
	if err != nil {
		return nil, err
	}

	v.Element = n
	v.ByteOrder = d.ByteOrder(n.endian)

	// If we have FixedValues, then check atleast one matches
	values := n.Values()
	log.Debugf("Number values: %s %s", n.values, values)

	if len(values) > 0 { // TODO Perhaps change with n.MustMatch()
		// Read the int value
		i, err := n.int(d.f, v)
		if err != nil {
			return nil, err
		}
		// Now check it matches one of the fixed values
		for _, fv := range values {
			if fv.value == i {
				log.Debug("matched %s", fv.value)
				return v, nil
			}
		}
		f, _ := n.format(i)
		log.Debugf("%v %s %s", n, n.Display(), f)
		return v, &validationError{e: n, msg: fmt.Sprintf("%q does match any of the fixed values %q", f, values)}
	}

	return v, nil
}


func (n *Custom) Read(d *Decoder) (*Value, error) {
	panic("TODO")
}

func (n *GrammarRef) Read(d *Decoder) (*Value, error) {
	panic("TODO")
}

func (n *Offset) Read(d *Decoder) (*Value, error) {
	panic("TODO")
}

func (n *ScriptElement) Read(d *Decoder) (*Value, error) {
	panic("TODO")
}

func (s *StructRef) Read(d *Decoder) (*Value, error) {
	log.Debugf("[%s] Read StructRef%s", d, s.Base.String())

	value, err := s.Structure().Read(d)
	if err != nil {
		return nil, err
	}
	value.Element = s
	return value, nil
}
