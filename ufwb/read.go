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

// Dummy Element, that is not actually found in the Grammar, but is used to represent padding
// through the file
var padElement = &Padding{Base: Base{"Padding", 0, "", ""}}

func (u *Ufwb) Read(d *Decoder) (*Value, error) {
	return d.read(u.Grammar)
}

func (g *Grammar) Read(d *Decoder) (*Value, error) {

	start, err := d.f.Tell()
	if err != nil {
		return nil, &validationError{e: g, err: err}
	}

	value := &Value{
		Offset:  start,
		Len:     0,
		Element: g,
	}

	// The start element may be repeated multiple times, so read via the elements.Read()
	elements := Elements([]Element{g.Start})
	return elements.Read(d, value, FixedOrder)
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

func (elements Elements) Read(d *Decoder, value *Value, order Order) (*Value, error) {

	parent := value.Element
	start := value.Offset

	// This Structure should not be bigger than the parent element
	bounds := d.ParentBounds()
	bounds_remaining := bounds.End - start

	if DEBUG && bounds.Start > start {
		panic(fmt.Sprintf("Starting before bounds %d < %d", start, bounds.Start))
	}

	childrenCount := make(map[ElementId]int64)

	i := 0
	eof := false

	log.Debugf("[0x%x] Starting %s (bounds: [0x%x-0x%x], max length: %d)", start, parent.IdString(), bounds.Start, bounds.End, bounds_remaining)

	// TODO If "value.Len < bounds_remaining" then we skip any tail elements that are zero length
	// such as certain scripts. If "value.Len <= bounds_remaining" then parsing breaks, which we
	// need to dig into more.
	for value.Len <= bounds_remaining && i < len(elements) {
		//log.Debugf("Loop %v, %v < %v, %v < %v", eof, childrenLength, length, i, len(elements))
		e := elements[i]

		max, err := d.eval(e.RepeatMax())
		if err != nil {
			return nil, &validationError{e: e, err: fmt.Errorf("RepeatMax eval failed: %s", err.Error())}
		}

		// If we have found the max of this element, move on
		if childrenCount[e] >= max {
			log.Debugf("Skipping %s found %d of %d", e.IdString(), childrenCount[e], max)
			i++
			continue
		}

		min, err := d.eval(e.RepeatMin())
		if err != nil {
			return nil, &validationError{e: e, err: fmt.Errorf("RepeatMin eval failed: %s", err.Error())}
		}

		// Ensure we parse this from the correct location.
		_, err = d.f.Seek(start+value.Len, io.SeekStart)
		if err != nil {
			return nil, &validationError{e: parent, err: err}
		}

		v, err := d.read(e)

		eof = isEof(err)

		// Only use the element if no error occurred (unless it was EOF)
		if v != nil && (err == nil || eof) {
			value.Children = append(value.Children, v)
			value.Len += v.Len
			childrenCount[v.Element]++

			// If we are variable order, start again from the first element for the next round
			if order == VariableOrder {
				log.Debugf("reset")
				i = 0
			}
		}

		if err != nil {
			// There was an error, but we can possibly try the next element.
			switch order {
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
					log.Debugf("[0x%x] Move on from: %s to: %s", start, elements[i].IdString(), elements[i+1].IdString())
				} else {
					log.Debugf("[0x%x] Move on from: %s to: end", start, elements[i].IdString())
				}
				i++

			default:
				return nil, &validationError{e: parent, err: fmt.Errorf("unknown order: %d", order)}
			}
		}
	}

	if parent.Length() != "" {
		log.Debugf("%s Loop %v, %v < %v, %v < %v", parent.IdString(), eof, value.Len, bounds_remaining, i, len(elements))

		// TODO Is this an error?
		if bounds_remaining > value.Len {
			log.Debugf("parent larger than children parent: %v, child: %v", bounds, value)
			padding := &Value{
				Offset:  start + value.Len,
				Len:     bounds_remaining - value.Len,
				Element: padElement,
			}
			value.Children = append(value.Children, padding)
			value.Len = bounds_remaining

			// TODO Eventually remove this error, since padding may be valid (but right now we are parsing strictly)
			panic(fmt.Sprintf("While developing we shouldn't need to add any padding! %v", padding))

		} else if value.Len > bounds_remaining {
			// The decoder ensures this shouldn't happen
			panic(fmt.Sprintf("children's length is greater than the parent length, %d vs %d", value.Len, bounds_remaining))
		}
	}

	for _, e := range elements {
		min, err := d.eval(e.RepeatMin())
		if err != nil {
			return nil, &validationError{e: e, err: fmt.Errorf("RepeatMin eval failed: %s", err.Error())}
		}

		if childrenCount[e] < min {
			return nil, &validationError{e: parent, err: fmt.Errorf("child %s repeated %d times, less than %d times required", e.IdString(), childrenCount[e], min)}
		}
	}

	if eof {
		return value, io.EOF
	}
	return value, nil
}

func (s *Structure) Read(d *Decoder) (*Value, error) {

	start, err := d.f.Tell()
	if err != nil {
		return nil, &validationError{e: s, err: err}
	}

	value := &Value{
		Offset:  start,
		Element: s,
	}

	//if start >= d.ParentBounds().End {
	//	return value, io.EOF
	//}

	/*
		// TODO This is trying to help tail element zero bytes long
		if s.Length() != "" {
			length, _ := d.Bits(s.Length(), s.LengthUnit()) // TODO check erro
			if length == 0 {
				return value, nil
			}
		}
	*/

	elements := Elements(s.Elements())
	return elements.Read(d, value, s.Order())
}

func (s *String) read(d *Decoder) (*Value, error) {
	start, err := d.f.Tell()
	if err != nil {
		return nil, err
	}

	end := d.ParentBounds().End
	maxLen := end - start

	var v *Value

	switch s.Typ() {
	case "zero-terminated", "delimiter-terminated":
		n, err := input.ReadUntil(d.f, s.delimiter, maxLen)
		if err != nil {
			return nil, err
		}
		v = &Value{Offset: start, Len: n, Element: s}

	case "fixed-length":
		v, err = lengthValue(d, s)
		if err != nil {
			return nil, err
		}

		// TODO Check for encoding.
		if _, err = input.ReadAndDiscard(d.f, v.Len); err != nil {
			return nil, err
		}

	case "pascal":
		// We assume 1 byte length prefix
		// TODO Write tests for this.
		// TODO Check this work for io.ErrUnexpectedEOF
		i, err := readInt(d.f, 1, false, binary.LittleEndian)
		if err != nil {
			return nil, err
		}

		length := int64(i.(uint8))

		// TODO Actually read it, and check for encoding.
		if _, err = input.ReadAndDiscard(d.f, length); err != nil {
			return nil, err
		}
		v = &Value{Offset: start, Len: length + 1, Element: s}

	default:
		return nil, fmt.Errorf("unknown string type %q", s.Typ())
	}

	// TODO Implement the fixed values
	//values := b.Values()
	//if len(values) > 0 && b.MustMatch().bool() {}

	return v, nil
}

func (s *String) Read(d *Decoder) (*Value, error) {
	v, err := s.read(d)
	if err != nil && err != io.EOF {
		err = &validationError{e: s, err: err}
	}
	return v, err
}

// lengthValue returns a value based on this Element's length field
// TODO Reconsider having this method, it seems to add little value
func lengthValue(d *Decoder, element Element) (*Value, error) {

	start, err := d.f.Tell()
	if err != nil {
		//return nil, &validationError{e: element, err: err}
		return nil, err
	}

	length, err := d.Bytes(element.Length(), element.LengthUnit())
	if err != nil {
		//return nil, &validationError{e: element, err: err}
		return nil, err
	}

	end := d.ParentBounds().End
	maxLen := end - start

	if maxLen == 0 {
		return nil, io.EOF
	}

	if length > maxLen {
		//return nil, &validationError{e: element, err: io.ErrUnexpectedEOF}
		return nil, io.ErrUnexpectedEOF
	}

	return &Value{
		Offset:  start,
		Len:     length,
		Element: element,
	}, nil
}

func (b *Binary) Read(d *Decoder) (*Value, error) {
	// TODO Binary.Read and Number.Read are almost identical.
	v, err := lengthValue(d, b)
	if err != nil {
		if err != io.EOF {
			err = &validationError{e: b, err: err}
		}
		return nil, err
	}

	bs, err := b.Bytes(d.f, v)
	if err != nil {
		return nil, &validationError{e: b, err: err}
	}

	// If we have FixedValues, then check at least one matches
	values := b.Values()
	if len(values) > 0 && b.MustMatch().bool() {
		// Now check it matches one of the fixed values
		for _, fv := range values {
			if bytes.Equal(fv.value, bs) {
				v.Extra = fv
				return v, nil
			}
		}

		f, err := b.format(bs)
		if err != nil {
			return v, &assertationError{e: b, err: fmt.Errorf("failed to format %v: %s", bs, err)}
		}

		formatedValues, err := b.formatValues()
		if err != nil {
			return v, &assertationError{e: b, err: fmt.Errorf("failed to format values %v: %s", values, err)}
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
	n, err := input.ReadFullAt(file, out, value.Offset)
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
		return 0, &assertationError{e: n, err: fmt.Errorf("reading value %v of another element", value)}
	}

	// Copy the value into a buffer first, because the ReaderAt interface is being used
	b := make([]byte, value.Len, value.Len)
	if _, err := input.ReadFullAt(file, b, value.Offset); err != nil {
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

// Uint returns the uint this file/value refers to. If the int doesn't fit into a uint64, it is truncated.
func (n *Number) Uint(file io.ReaderAt, value *Value) (uint64, error) {
	i, err := n.int(file, value)
	if err != nil {
		return 0, err
	}
	if n.Signed() {
		return uint64(reflect.ValueOf(i).Int()), nil
	} else {
		return reflect.ValueOf(i).Uint(), nil
	}
}

func (n *Number) Read(d *Decoder) (*Value, error) {
	v, err := lengthValue(d, n)
	if err != nil {
		if err != io.EOF {
			err = &validationError{e: n, err: err}
		}
		return nil, err
	}

	v.ByteOrder = d.ByteOrder(n.Endian())

	// Read the int value
	i, err := n.int(d.f, v)
	if err != nil {
		return nil, err // n.int returns validationError so no need to wrap
	}

	// If we have FixedValues, then check atleast one matches
	values := n.Values()
	if len(values) > 0 && n.MustMatch().bool() {
		// Now check it matches one of the fixed values
		for _, fv := range values {
			if intEqual(fv.value, i) {
				v.Extra = fv
				return v, nil
			}
		}
		f, err := n.format(i)
		if err != nil {
			return v, &assertationError{e: n, err: fmt.Errorf("failed to format %v: %s", i, err)}
		}

		formatedValues, err := n.formatValues()
		if err != nil {
			return v, &assertationError{e: n, err: fmt.Errorf("failed to format values %v: %s", values, err)}
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

func (s *Script) Read(d *Decoder) (*Value, error) {
	start, err := d.f.Tell()
	if err != nil {
		return nil, err
	}

	// TODO Don't have a switch, instead have different Script types
	switch s.Language() {
	case "lua", "Lua":
		err = s.RunLua(d)
	default:
		return nil, fmt.Errorf("unsupported langauge %q", s.Language())
	}

	// Create a empty value
	return &Value{
		Offset:  start,
		Len:     0,
		Element: s,
	}, err
}

func (p *Padding) Read(d *Decoder) (*Value, error) {
	// TODO We never should read a padding element
	panic("This should never be called")
}

func (s *StructRef) Read(d *Decoder) (*Value, error) {
	v, err := d.read(s.Structure())
	if err != nil {
		return nil, err
	}
	v.Element = s

	return v, nil
}
