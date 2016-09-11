package ufwb

// This file uses the grammar to parse the binary file.

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"strconv"

	log "github.com/Sirupsen/logrus"
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

// File interface provides the minimum needed to parse the binary file.
type File interface {
	io.Seeker

	io.Reader
	io.ReaderAt
	io.ByteReader

	Tell() (int64, error) // Here for convinence, perhaps remove.
}

// Value represents one of the parsed elements in the file.
// It doesn't contain the element, just the offset where it starts, and which element it is.
type Value struct {
	Offset  int64 // In bytes from the beginning of the file
	Len     int64 // In bytes
	Element Element

	Children  []*Value
	ByteOrder binary.ByteOrder // Only used for Number, TODO, and TODO
}

func (v *Value) Name() string {
	return v.Element.GetName()
}

func (v *Value) Description() string {
	return v.Element.GetDescription()
}

// String returns this value's string representation (based on display, etc)
func (v *Value) Format(f File) (string, error) {
	return v.Element.Format(f, v)
}

// Read returns the string representation of this value
func (v *Value) Read(f File) (string, error) {
	return "", nil
}

func (v *Value) Write(f File) {
	panic("TODO")
}

// Validiate checks if this Value is valid, mainly used for debugging.
func (v *Value) Validiate() error {
	if v == nil {
		return fmt.Errorf("nil Value")
	}
	if v.Len <= 0 {
		return fmt.Errorf("Len = %d want >0", v.Len)
	}
	if v.Offset < 0 {
		return fmt.Errorf("Offset = %d want >0", v.Offset)
	}
	if v.Element == nil {
		return fmt.Errorf("Element = nil want a valid value", v.Element)
	}

	end := v.Offset + v.Len
	for _, child := range v.Children {
		if err := child.Validiate(); err != nil {
			return err
		}
		if child.Offset < v.Offset || (child.Offset+child.Len) > end {
			return fmt.Errorf("child Value outside of parents bounds: %v > %v", v, child)
		}
	}

	return nil
}

func (d *Decoder) Decode() (*Value, error) {
	return d.u.Read(d)
}

func (u *Ufwb) Read(d *Decoder) (*Value, error) {
	log.Debugln("Ufwb Read")
	return u.Grammar.Read(d)
}

func (g *Grammar) Read(d *Decoder) (*Value, error) {
	log.Debugf("Read Grammar%s", g.Base.String())
	return g.Start.Read(d)
}

func (g *Grammar) Format(f File, value *Value) (string, error) {
	return g.Start.Format(f, value)
}

func (s *Structure) Read(d *Decoder) (*Value, error) {
	log.Debugf("Read Structure%s", s.Base.String())

	start, err := d.f.Tell()
	if err != nil {
		return nil, err
	}

	length := int64(0)
	var children []*Value

	if s.Order() == FixedOrder {
		for _, e := range s.Elements() {
			v, err := e.Read(d)
			if err != nil {
				return nil, err
			}
			length += v.Len
			children = append(children, v)
		}
	} else if s.Order() == VariableOrder {
		// TODO FIX:
		for _, e := range s.Elements() {
			v, err := e.Read(d)
			if err != nil {
				return nil, err
			}
			length += v.Len
			children = append(children, v)
		}
	} else {
		return nil, &validationError{e: s, msg: fmt.Sprintf("unknown order: %s", s.Order())}
	}

	return &Value{
		Offset:   start,
		Len:      length,
		Element:  s,
		Children: children,
	}, nil
}

func (n *Structure) Format(f File, value *Value) (string, error) {
	panic("TODO")
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
	log.Debugf("Read String%s", s.Base.String())

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
	log.Debugf("Read Binary%s", b.Base.String())

	v, err := skip(d, b.Length(), b.lengthUnit)
	if err != nil {
		return nil, err
	}

	v.Element = b
	return v, nil
}

func (n *String) Format(f File, value *Value) (string, error) {
	panic("TODO")
}

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
	return i, err
}

// Int returns the int this file/value refers to.
func (n *Number) Int(f File, value *Value) (int64, error) {
	i, err := n.int(f, value)
	if err != nil {
		return 0, err
	}
	return reflect.ValueOf(i).Elem().Int(), nil
}

func (n *Number) Uint(f File, value *Value) (uint64, error) {
	i, err := n.int(f, value)
	if err != nil {
		return 0, err
	}
	return reflect.ValueOf(i).Elem().Uint(), nil
}

func (n *Number) Format(f File, value *Value) (string, error) {
	base := n.display.Base()
	if base == 0 {
		return "", fmt.Errorf("invalid base %d", base)
	}
	if n.signed {
		i, err := n.Int(f, value)
		return strconv.FormatInt(i, base), err
	} else {
		i, err := n.Uint(f, value)
		return strconv.FormatUint(i, base), err
	}
}

func (n *Number) Read(d *Decoder) (*Value, error) {
	log.Debugf("Read Number%s", n.Base.String())

	v, err := skip(d, n.Length(), n.LengthUnit())
	if err != nil {
		return nil, err
	}

	v.Element = n
	v.ByteOrder = d.ByteOrder(n.endian)

	// If we have FixedValues, then check atleast one matches
	values := n.Values()
	if len(values) > 0 { // TODO Perhaps change with n.MustMatch()
		// Read the int value
		i, err := n.int(d.f, v)
		if err != nil {
			return nil, err
		}
		// Now check it matches one of the fixed values
		for _, fv := range values {
			if fv.value == i {
				return v, nil
			}
		}
		return v, &validationError{e: n, msg: fmt.Sprintf("%d does match any of the fixed values", i)}
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
	log.Debugf("Read StructRef%s", s.Base.String())

	value, err := s.Structure().Read(d)
	if err != nil {
		return nil, err
	}
	value.Element = s
	return value, nil
}

func (n *Binary) Format(f File, value *Value) (string, error) {
	panic("TODO")
}

func (n *Custom) Format(f File, value *Value) (string, error) {
	panic("TODO")
}

func (n *GrammarRef) Format(f File, value *Value) (string, error) {
	panic("TODO")
}

func (n *Offset) Format(f File, value *Value) (string, error) {
	panic("TODO")
}

func (n *ScriptElement) Format(f File, value *Value) (string, error) {
	panic("TODO")
}

func (n *StructRef) Format(f File, value *Value) (string, error) {
	return n.Structure().Format(f, value)
}

/*
func (n *Mask) Format(f File, value *Value) (string, error) {
	panic("TODO")
}

func (n *FixedValues) Format(f File, value *Value) (string, error) {
	panic("TODO")
}

func (n *FixedValue) Format(f File, value *Value) (string, error) {
	panic("TODO")
}

func (n *Script) Format(f File, value *Value) (string, error) {
	panic("TODO")
}
*/
