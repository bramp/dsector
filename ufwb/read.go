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
	"strings"
	"github.com/pkg/errors"
	"math"
)

type startElement struct {
	start int64
	element Element
}

const DEBUG = true
const MAX_STACK = 10

type StackPrinter []startElement

func (stack StackPrinter) String() string {
	var buffer bytes.Buffer

	for _, s := range stack {
		buffer.WriteString(fmt.Sprintf("[%d %s] ", s.start, s.element.String()))
	}

	return buffer.String()
}

type Decoder struct {
	u             *Ufwb
	f             File

	// dynamicEndian be changed by scripts during processing.
	dynamicEndian binary.ByteOrder

	stack         []startElement

	prevMap       map[string]*Value
}

func NewDecoder(u *Ufwb, f File) *Decoder {
	return &Decoder{
		u: u,
		f: f,
		prevMap: make(map[string]*Value),
	}
}

func (d *Decoder) read(e Element) (*Value, error) {
	start, err := d.f.Tell()
	if err != nil {
		return nil, err
	}

	if len(d.stack) > MAX_STACK {
		panic(fmt.Sprintf("Exceeded max parsing stack depth of %d", MAX_STACK))
	}

	log.Debugf("[%d] Reading: %s", start, e.String())
	//log.Debugf("[%d] Stack: %s", start, StackPrinter(d.stack))

	d.stack = append(d.stack, startElement{start, e})
	v, err := e.Read(d)
	log.Debugf("[%d] Read: %s %s", start, v, err)

	d.stack = d.stack[:len(d.stack) - 1]

	if v != nil {
		if DEBUG {
			// Debug / validation code
			v.mustValidiate()
			if pos, err := d.f.Tell(); err != nil {
				if (v.Offset + v.Len) != pos {
					panic(fmt.Sprintf("Decoder was not left at right position after %v", v))
				}
			}
		}

		d.prevMap[v.Element.Name()] = v // TODO I'm not sure if this is the best defintion of "last"
	}

	return v, err
}

// TODO Make this actually eval the string, and determine the right value
func (d *Decoder) eval(r Reference) (int64, error) {

	str := string(r)
	if str == "remaining" {
		return d.remaining()
	}

	if str == "unlimited" {
		return math.MaxInt64, nil
	}

	if strings.HasPrefix(str, "prev.") {
		return d.prev(str)
	}

	// Try a number
	i, err := strconv.Atoi(str)
	if err != nil {
		panic(err) // PANIC While we debug how eval should work. Eventually return error
	}
	return int64(i), err
}

// currentStruct returns the most recent structure on the stack
func (d *Decoder) currentStruct() (int64, *Structure, error) {
	for i := len(d.stack) - 1; i >= 0; i-- {
		if s, ok := d.stack[i].element.(*Structure); ok {
			return d.stack[i].start, s, nil
		}
	}
	return -1, nil, errors.New("No structure found. This should never happen")
}

// remaining returns the number of bytes remaining in the current structure.
func (d *Decoder) remaining() (int64, error) {
	pos, err := d.f.Tell()
	if err != nil {
		return -1, err
	}

	start, s, err := d.currentStruct()
	if err != nil {
		return -1, err
	}

	len, err := d.eval(s.Length())
	if err != nil {
		return -1, err
	}

	return (start - pos + len), nil
}

// prev returns the value read by the previous element of this name.
func (d *Decoder) prev(name string) (int64, error) {
	name = strings.TrimPrefix(name, "prev.")

	// TODO Instead of using a map, recurse up the d.stack/tree

	v, ok := d.prevMap[name]
	if !ok {
		return -1, fmt.Errorf("no previous element named %q found", name)
	}

	n, ok := v.Element.(*Number)
	if !ok {
		return -1, fmt.Errorf("previous element %q must be a Number", name)
	}

	return n.Int(d.f, v)
}


// ByteOrder returns the current byte order
// TODO Delete this method
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
	panic("blah")
}

// File interface provides the minimum needed to parse the binary file.
type File interface {
	io.Seeker

	io.Reader
	io.ReaderAt
	io.ByteReader

	Tell() (int64, error) // Here for convenience, perhaps remove.
}

func (d *Decoder) Decode() (*Value, error) {
	return d.u.Read(d)
}

func (u *Ufwb) Read(d *Decoder) (*Value, error) {
	return u.Grammar.Read(d)
}

func (g *Grammar) Read(d *Decoder) (*Value, error) {
	return g.Start.Read(d)
}

func (s *Structure) Read(d *Decoder) (*Value, error) {
	start, err := d.f.Tell()
	if err != nil {
		return nil, &validationError{e: s, msg: err.Error()}
	}

	if DEBUG && start > 10000 {
		panic("DEBUG: ENDING EARLY")
	}

	length := int64(math.MaxInt64)
	if s.Length() != "" {
		length, err = d.eval(s.Length())
		if err != nil {
			return nil, &validationError{e: s, msg: err.Error()}
		}
	}

	var eof error
	var children []*Value
	var childrenLength int64
	childrenCount := make(map[Element]int64)

	elements := s.Elements() // TODO This doesn't work correctly with extends
	i := 0
	for eof == nil && childrenLength < length && i < len(elements) {
		e := elements[i]

		max, err := d.eval(e.RepeatMax())
		if err != nil {
			return nil, &validationError{e: s, msg: fmt.Sprintf("RepeatMax eval failed: %s", err.Error())}
		}

		if max > childrenCount[e] {
			i++
			continue
		}

		log.Debugf("Looking at %d %s", i, e)
		_, err = d.f.Seek(start + childrenLength, io.SeekStart)
		if err != nil {
			return nil, &validationError{e: s, msg: err.Error()}
		}

		v, err := d.read(e)
		if v != nil {
			str, _ := v.Format(d.f)
			log.Debugf("Found: %s %s %v", v, e, str)

			childrenLength += v.Len
			children = append(children, v)
			childrenCount[v.Element]++

			switch s.Order() {
				case VariableOrder:
					log.Debugf("variable order start again")
					// If we are variable order, start again from the first element for the next round
					i = 0
					// TODO Detect if we are in parsing loop
				case FixedOrder:
					// Have we found the minimum number of this element, if so move on.
					min, err := d.eval(e.RepeatMin())
					if err != nil {
						return nil, &validationError{e: s, msg: fmt.Sprintf("RepeatMin eval failed: %s", err.Error())}
					}
					if childrenCount[e] >= min {
						i++
					}
			}
		}

		if err != nil {
			if err == io.EOF {
				// TODO Check if an end was expected here
				eof = io.EOF
				break
			}

			switch s.Order() {
			case FixedOrder:
				return nil, err

			case VariableOrder:
				log.Debugf("try another %v: %s", e, err)

				// This one failed, try another element
				i++

			default:
				return nil, &validationError{e: s, msg: fmt.Sprintf("unknown order: %s", s.Order())}
			}
		}
	}

	// TODO Check if we reached the min children

	if s.Length() != "" {
		// TODO Is this an error?
		if length > childrenLength {
			padding := &Value{
				Offset:   start + childrenLength,
				Len:      length - childrenLength,
				Element:  nil, // TODO maybe change to a padding element
			}
			children = append(children, padding)
			childrenLength = length

		} else if length < childrenLength {
			// TODO Is this an error?
			return nil, &validationError{
				e: s,
				msg: fmt.Sprintf("children length is greater than the structure length, %d vs %s", childrenLength, length),
			}
		}
	}

	return &Value{
		Offset:   start,
		Len:      childrenLength,
		Element:  s,
		Children: children,
	}, eof
}

func (s *String) Read(d *Decoder) (*Value, error) {
	start, err := d.f.Tell()
	if err != nil {
		return nil, &validationError{e: s, msg: err.Error()}
	}

	var v *Value

	switch s.Typ() {
	case "zero-terminated":
		n, err := seekUntil(d.f, '\x00')
		if err != nil {
			return nil, &validationError{e: s, msg: err.Error()}
		}
		v = &Value{Offset: start, Len: n, Element: s}

	case "fixed-length":
		len, err := d.eval(s.Length())
		if err != nil {
			return nil, &validationError{e: s, msg: err.Error()}
		}
		d.f.Seek(len, io.SeekCurrent)
		v = &Value{Offset: start, Len: len, Element: s}

	default:
		return nil, fmt.Errorf("unknown type %q", s.Typ())
	}

	return v, nil
}

// skip moves the decoder forward by length, returning a Value that covers the range
func skip(d *Decoder, length Reference, lengthUnit LengthUnit) (*Value, error) { // TODO Make this accept a "Length type", which combines the unit and the value
	start, err := d.f.Tell()
	if err != nil {
		return nil, err
	}

	len, err := d.eval(length)
	if err != nil {
		return nil, err
	}

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
	// TODO Binary.Read and Number.Read are almost identical.
	v, err := skip(d, b.Length(), b.LengthUnit())
	if err != nil {
		return nil, &validationError{e: b, msg: err.Error()}
	}

	v.Element = b

	// If we have FixedValues, then check at least one matches
	values := b.Values()
	if len(values) > 0 && b.MustMatch().bool() {
		// Read the bytes value
		bs, err := b.Bytes(d.f, v)
		if err != nil {
			return nil, &validationError{e: b, msg: err.Error()}
		}
		// Now check it matches one of the fixed values
		for _, fv := range values {
			if bytes.Equal(fv.value, bs) {
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
			e: b,
			msg: fmt.Sprintf("%q does match any of the fixed values %q", f, formatedValues),
		}
	}

	return v, nil
}

// Bytes returns the bytes from file, found at Value
func (b *Binary) Bytes(file File, value *Value) ([]byte, error) {
	out := make([]byte, value.Len, value.Len)
	log.Debugf("READING AT %d", value.Offset)
	n, err := file.ReadAt(out, value.Offset)
	if err != nil {
		return nil, &validationError{e: b, msg: err.Error()}
	}
	return out[:n], nil

}

// int returns the integer stored at Value in f. The returned
// integer is one of int{8,16,32,64} or uint{8,16,32,64} depending
// on the width and sign of the integer.
func (n *Number) int(f File, value *Value) (interface{}, error) {
	if _, err := f.Seek(value.Offset, io.SeekStart); err != nil {
		return 0, &validationError{e: n, msg: err.Error()}
	}

	i, err := readInt(f, value.Len, n.Signed(), value.ByteOrder)
	if err != nil && err != io.EOF {
		return 0, &validationError{e: n, msg: err.Error()}
	}

	return i, err
}

// Int returns the int this file/value refers to. If the int doesn't fit into a int64, it is truncated.
func (n *Number) Int(f File, value *Value) (int64, error) {
	i, err := n.int(f, value)
	if err != nil {
		return 0, err
	}
	if n.Signed() {
		return reflect.ValueOf(i).Int(), nil
	} else {
		return int64(reflect.ValueOf(i).Uint()), nil
	}
}

/*
func (n *Number) Uint(f File, value *Value) (uint64, error) {
	i, err := n.int(f, value)
	if err != nil {
		return 0, err
	}
	return reflect.ValueOf(i).Uint(), nil
}
*/


func (n *Number) Read(d *Decoder) (*Value, error) {
	v, err := skip(d, n.Length(), n.LengthUnit())
	if err != nil {
		return nil, &validationError{e: n, msg: err.Error()}
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
			e: n,
			msg: fmt.Sprintf("%q does match any of the fixed values %q", f, formatedValues),
		}
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

	v, err := d.read(s.Structure())
	if err != nil {
		return nil, err
	}
	v.Element = s

	return v, nil
}
