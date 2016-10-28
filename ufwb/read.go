package ufwb

// This file uses the grammar to parse the binary file.

import (
	"fmt"
	"io"
	"reflect"

	"bramp.net/dsector/input"
	"bytes"
	"encoding/binary"
	log "github.com/Sirupsen/logrus"
)

var padElement = &Padding{Base: Base{"Padding", 0, "", ""}}

func (u *Ufwb) Read(d *Decoder) (*Value, error) {
	return d.read(u.Grammar)
}

func (g *Grammar) Read(d *Decoder) (*Value, error) {
	s, err := d.read(g.Start)
	if err != nil {
		return nil, err
	}

	return &Value{
		Offset:   s.Offset,
		Len:      s.Len,
		Element:  g,
		Children: []*Value{s},
	}, nil
}

// isEof returns if this error represents the end of file
func isEof(err error) bool {
	if err == io.EOF {
		return true
	}

	if e, ok := err.(Eof); ok {
		return e.IsEof()
	}

	return false
}

func (s *Structure) Read(d *Decoder) (*Value, error) {
	start, err := d.f.Tell()
	if err != nil {
		return nil, &validationError{e: s, err: err}
	}

	bounds := d.ParentBounds()
	length := bounds.End - start

	if DEBUG && bounds.Start > start {
		panic(fmt.Sprintf("Starting before bounds %d < %d", start, bounds.Start))
	}

	childrenCount := make(map[ElementLite]int64)

	value := &Value{
		Offset:  start,
		Element: s,
	}

	i := 0
	elements := s.Elements() // TODO This doesn't work correctly with extends
	eof := false

	log.Debugf("[%d] Starting %s (bounds: [%d-%d], max length: %d)", start, s.IdString(), bounds.Start, bounds.End, length)

	for value.Len < length && i < len(elements) {
		//log.Debugf("Loop %v, %v < %v, %v < %v", eof, childrenLength, length, i, len(elements))

		e := elements[i]

		max, err := d.eval(e.RepeatMax())
		if err != nil {
			return nil, &validationError{e: s, err: fmt.Errorf("RepeatMax eval failed: %s", err.Error())}
		}

		// If we have found the max of this element, move on
		if childrenCount[e] >= max {
			log.Debugf("Skipping %s found %d of %d", e.IdString(), childrenCount[e], max)
			i++
			continue
		}

		min, err := d.eval(e.RepeatMin())
		if err != nil {
			return nil, &validationError{e: s, err: fmt.Errorf("RepeatMin eval failed: %s", err.Error())}
		}

		// Ensure we parse this from the correct location.
		_, err = d.f.Seek(start+value.Len, io.SeekStart)
		if err != nil {
			return nil, &validationError{e: s, err: err}
		}

		v, err := d.read(e)

		eof = isEof(err)

		// Only use the element if no error occurred (unless it was EOF)
		if v != nil && (err == nil || eof) {
			value.Children = append(value.Children, v)
			value.Len += v.Len
			childrenCount[v.Element]++

			// If we are variable order, start again from the first element for the next round
			if s.Order() == VariableOrder {
				log.Debugf("reset")
				i = 0
			}
		}

		if err != nil {
			// There was an error, but we can possibly try the next element.
			switch s.Order() {
			case FixedOrder:
				// There was an error reading this item, and we have already found the minimum
				// number of this element, so lets move on.
				if childrenCount[e] >= min {
					i++
					continue
				}

				// Otherwise return the error
				return nil, err

			case VariableOrder:
				// This one failed, try another element
				if i < len(elements)-1 {
					log.Debugf("move on from: %s to: %s", elements[i].IdString(), elements[i+1].IdString())
				} else {
					log.Debugf("move on from: %s to: end", elements[i].IdString())
				}
				i++

			default:
				return nil, &validationError{e: s, err: fmt.Errorf("unknown order: %s", s.Order())}
			}
		}
	}

	// TODO Check if we reached the min children
	// TODO Check if we read all children

	if s.Length() != "" {
		log.Debugf("%s Loop %v, %v < %v, %v < %v", s.IdString(), eof, value.Len, length, i, len(elements))

		// TODO Is this an error?
		if length > value.Len {
			padding := &Value{
				Offset:  start + value.Len,
				Len:     length - value.Len,
				Element: padElement,
			}
			value.Children = append(value.Children, padding)
			value.Len = length

			// TODO remove this panic
			panic(fmt.Sprintf("Shouldn't need to add any padding! %v", padding))

		} else if length < value.Len {
			panic(fmt.Sprintf("children length is greater than the structure length, %d vs %d", value.Len, length))
		}
	}

	if eof {
		return value, io.EOF
	}
	return value, nil
}

func (s *String) Read(d *Decoder) (*Value, error) {
	start, err := d.f.Tell()
	if err != nil {
		return nil, &validationError{e: s, err: err}
	}

	var v *Value

	switch s.Typ() {
	case "zero-terminated":
		n, err := input.SeekUntil(d.f, '\x00', d.ParentBounds().End-start)
		if err != nil {
			return nil, &validationError{e: s, err: err}
		}
		v = &Value{Offset: start, Len: n, Element: s}

	case "fixed-length":
		length, err := d.eval(s.Length()) // TODO Do I need this, or can I use ParentBounds()?
		if err != nil {
			return nil, &validationError{e: s, err: err}
		}
		d.f.Seek(length, io.SeekCurrent) // Seek beyond the string (as if we read it)
		v = &Value{Offset: start, Len: length, Element: s}

	case "pascal":
		// We assume 1 byte length prefix
		// TODO Write tests for this.
		i, err := readInt(d.f, 1, false, binary.LittleEndian)
		if err != nil {
			return nil, &validationError{e: s, err: err}
		}

		length := int64(i.(uint8))

		d.f.Seek(length, io.SeekCurrent) // Seek beyond the string (as if we read it)
		v = &Value{Offset: start, Len: length + 1, Element: s}

	default:
		return nil, fmt.Errorf("unknown string type %q", s.Typ())
	}

	// TODO Implement the fixed values
	//values := b.Values()
	//if len(values) > 0 && b.MustMatch().bool() {

	return v, nil
}

// skip moves the decoder forward by length, returning a Value that covers the range
func skip(d *Decoder, length Reference, lengthUnit LengthUnit) (*Value, error) { // TODO Make this accept a "Length type", which combines the unit and the value
	start, err := d.f.Tell()
	if err != nil {
		return nil, err
	}

	len, err := d.Bytes(length, lengthUnit)
	if err != nil {
		return nil, err
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
	// TODO Binary.Read and Number.Read are almost identical.
	v, err := skip(d, b.Length(), b.LengthUnit())
	if err != nil {
		return nil, &validationError{e: b, err: err}
	}

	v.Element = b

	// If we have FixedValues, then check at least one matches
	values := b.Values()
	if len(values) > 0 && b.MustMatch().bool() {
		// Read the bytes value
		bs, err := b.Bytes(d.f, v)
		if err != nil {
			return nil, &validationError{e: b, err: err}
		}
		// Now check it matches one of the fixed values
		for _, fv := range values {
			if bytes.Equal(fv.value, bs) {
				v.Extra = fv
				return v, nil
			}
		}

		f, err := b.format(bs)
		if err != nil {
			return v, &assertationError{e: b, msg: fmt.Sprintf("failed to format %v: %s", bs, err)}
		}

		formatedValues, err := b.formatValues()
		if err != nil {
			return v, &assertationError{e: b, msg: fmt.Sprintf("failed to format values %v: %s", values, err)}
		}

		return nil, &validationError{
			e:   b,
			err: fmt.Errorf("%q does match any of the fixed values %q", f, formatedValues),
		}
	}

	return v, nil
}

// Bytes returns the bytes from file, found at Value
func (b *Binary) Bytes(file io.ReaderAt, value *Value) ([]byte, error) {
	out := make([]byte, value.Len, value.Len)
	n, err := file.ReadAt(out, value.Offset)
	if err != nil {
		return nil, &validationError{e: b, err: err}
	}
	return out[:n], nil

}

// int returns the integer stored at Value in f. The returned
// integer is one of int{8,16,32,64} or uint{8,16,32,64} depending
// on the width and sign of the integer.
func (n *Number) int(file io.ReaderAt, value *Value) (interface{}, error) {
	if n != value.Element {
		return 0, &assertationError{e: n, err: fmt.Errorf("trying to use wrong value %v", value)}
	}

	b := make([]byte, value.Len, value.Len)
	if _, err := file.ReadAt(b, value.Offset); err != nil {
		return 0, &validationError{e: n, err: err}
	}

	r := bytes.NewReader(b)
	i, err := readInt(r, value.Len, n.Signed(), value.ByteOrder)
	if err != nil && err != io.EOF { // TODO Remove EOF special case
		return 0, &validationError{e: n, err: err}
	}

	return i, err
}

// Int returns the int this file/value refers to. If the int doesn't fit into a int64, it is truncated.
func (n *Number) Int(file io.ReaderAt, value *Value) (int64, error) {
	i, err := n.int(file, value)
	if err != nil {
		return 0, err
	}
	if n.Signed() {
		return reflect.ValueOf(i).Int(), nil
	} else {
		return int64(reflect.ValueOf(i).Uint()), nil
	}
}

func (n *Number) Read(d *Decoder) (*Value, error) {
	v, err := skip(d, n.Length(), n.LengthUnit())
	if err != nil {
		return nil, &validationError{e: n, err: err}
	}

	v.Element = n
	v.ByteOrder = d.ByteOrder(n.Endian())

	// If we have FixedValues, then check atleast one matches
	values := n.Values()
	if len(values) > 0 && n.MustMatch().bool() {
		// Read the int value
		i, err := n.int(d.f, v)
		if err != nil {
			return nil, err // n.int returns validationError so no need to wrap
		}
		// Now check it matches one of the fixed values
		for _, fv := range values {
			if intEqual(fv.value, i) {
				v.Extra = fv
				return v, nil
			}
		}
		f, err := n.format(i)
		if err != nil {
			return v, &assertationError{e: n, msg: fmt.Sprintf("failed to format %v: %s", i, err)}
		}

		formatedValues, err := n.formatValues()
		if err != nil {
			return v, &assertationError{e: n, msg: fmt.Sprintf("failed to format values %v: %s", values, err)}
		}

		return v, &validationError{
			e:   n,
			err: fmt.Errorf("%q does match any of the fixed values %q", f, formatedValues),
		}
	}

	return v, nil
}

func (c *Custom) Read(d *Decoder) (*Value, error) {
	panic("TODO")
}

func (g *GrammarRef) Read(d *Decoder) (*Value, error) {
	panic("TODO")
}

func (o *Offset) Read(d *Decoder) (*Value, error) {
	panic("TODO")
}

func (s *ScriptElement) Read(d *Decoder) (*Value, error) {
	panic("TODO")
}

func (p *Padding) Read(d *Decoder) (*Value, error) {
	// TODO We never should read a padding element
	panic("TODO")
}

func (s *StructRef) Read(d *Decoder) (*Value, error) {
	v, err := d.read(s.Structure())
	if err != nil {
		return nil, err
	}
	v.Element = s

	return v, nil
}
