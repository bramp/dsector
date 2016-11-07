package ufwb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"bramp.net/dsector/input"
	"errors"
	log "github.com/Sirupsen/logrus"
	"io"
	"strconv"
)

const DEBUG = true
const MAX_STACK = 10

type ElementBounds struct {
	Element Element
	Start   int64
	End     int64
}

func (bounds *ElementBounds) Length() int64 {
	return bounds.End - bounds.Start
}

func (bounds *ElementBounds) String() string {
	return fmt.Sprintf("[0x%x-0x%x] %s", bounds.Start, bounds.End, bounds.Element.IdString())
}

type StackPrinter []ElementBounds

func (stack StackPrinter) String() string {
	var buffer bytes.Buffer

	for _, s := range stack {
		buffer.WriteString(fmt.Sprintf("[%d %s] ", s.Start, s.Element.IdString()))
	}

	return buffer.String()
}

type Decoder struct {
	u *Ufwb
	f input.Input

	stack  []ElementBounds
	values []*Value

	// dynamicEndian be changed by scripts during processing.
	dynamicEndian binary.ByteOrder

	// debugFunc hooks a "debug" function into the script env
	debugFunc func(interface{})
}

func NewDecoder(u *Ufwb, f input.Input) *Decoder {
	return &Decoder{
		u: u,
		f: f,
	}
}

func (d *Decoder) Decode() (*Value, error) {
	d.stack = nil
	d.values = nil
	return d.u.Read(d)
}

func (d *Decoder) ParentBounds() *ElementBounds {
	if len(d.stack) > 0 {
		return &d.stack[len(d.stack)-1]
	}

	return nil
}

func (d *Decoder) read(e Element) (*Value, error) {

	if len(d.stack) > MAX_STACK {
		log.Debugf("%s", StackPrinter(d.stack))
		panic(fmt.Sprintf("Exceeded max parsing stack depth of %d", MAX_STACK))
	}

	start, err := d.f.Tell()
	if err != nil {
		return nil, err
	}
	var end int64

	log.Debugf("[0x%x] Reading: %s", start, e.IdString())
	//log.Debugf("[0x%x] Stack: %s", start, StackPrinter(d.stack))

	bounds := d.ParentBounds()
	if bounds == nil {
		// Seek to get the size of this file
		end, err = d.f.Seek(0, io.SeekEnd)
		if err != nil {
			return nil, err
		}

		// Reset to beginning
		_, err = d.f.Seek(start, io.SeekStart)
		if err != nil {
			return nil, err
		}

	} else {
		end = bounds.End
	}

	if start == end {
		return nil, io.EOF
	}

	// If the element has a smaller length, then bound it.
	if e.Length() != "" {
		length, err := d.eval(e.Length())
		if err != nil {
			return nil, &validationError{e: e, err: err}
		}
		if length < (end - start) {
			end = start + length
		}
	}

	d.stack = append(d.stack, ElementBounds{
		Element: e,
		Start:   start,
		End:     end,
	})

	// Real parsing is in here
	v, err := e.Read(d)

	vformat := ""
	if v != nil {
		vformat, _ = v.Format(d.f)
	}
	log.Debugf("[0x%x] Read: %s %s %q err:%v", start, e.IdString(), v, ellipsis(vformat, 10), err)

	// Pop off the stack
	d.stack = d.stack[:len(d.stack)-1]

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

		d.values = append(d.values, v)
	}

	return v, err
}

// TODO Make this actually eval the string, and determine the right value
func (d *Decoder) eval(r Reference) (i int64, err error) {
	str := string(r)

	switch {
	case str == "remaining":
		i, err = d.remaining()

	case str == "unlimited":
		i = math.MaxInt64

	case strings.HasPrefix(str, "prev."):
		name := strings.TrimPrefix(str, "prev.")

		v, err := d.prevByName(name)
		if err != nil {
			return -1, err
		}

		n, ok := v.Element.(*Number)
		if !ok {
			return -1, fmt.Errorf("previous element %q must be a Number", name)
		}

		log.Debugf("prev(%q) value: %s", name, v)
		return n.Int(d.f, v)

	default:
		// Try a number // TODO Remove this path when we created ConstReferences
		i, err = strconv.ParseInt(str, 10, 0)
		if err != nil {
			panic(err) // PANIC While we debug how eval should work. Eventually return error
		}
	}

	log.Debugf("eval(%q) = %d, %v", str, i, err)
	return
}

// currentStruct returns the most recent structure on the stack
func (d *Decoder) currentStruct() (int64, *Structure, error) {
	for i := len(d.stack) - 1; i >= 0; i-- {
		if s, ok := d.stack[i].Element.(*Structure); ok {
			return d.stack[i].Start, s, nil
		}
	}
	return -1, nil, errors.New("No structure found. This should never happen")
}

// remaining returns the number of bytes remaining in the current structure.
func (d *Decoder) remaining() (int64, error) {

	// TODO change this to use d.ParentBounds()

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

// prev returns the previous value.
func (d *Decoder) prev() (*Value, error) {
	if len(d.values) > 0 {
		return d.values[len(d.values)-1], nil
	}

	return nil, errors.New("no previous element")
}

// prevByName returns the value read by the previous element of this name.
func (d *Decoder) prevByName(name string) (*Value, error) {

	for i := len(d.values) - 1; i >= 0; i-- {
		if d.values[i].Element.Name() == name {
			return d.values[i], nil
		}
	}

	return nil, fmt.Errorf("no previous element named %q found", name)
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

// Bytes returns the number of bytes this length represents.
// If the unit is in bits, it rounds up or down.
func (d *Decoder) Bytes(length Reference, unit LengthUnit) (int64, error) {
	len, err := d.eval(length)
	if err != nil {
		return -1, err
	}

	switch unit {
	case BitLengthUnit:
		if len%8 != 0 {
			return -1, fmt.Errorf("unsupported length %s = %d bits", length, len)
		}
		return len / 8, nil
	case ByteLengthUnit:
		return len, nil
	}

	return -1, fmt.Errorf("unknown length unit %d", unit)
}

// Bits returns the number of bits this length represents.
func (d *Decoder) Bits(length Reference, unit LengthUnit) (int64, error) {
	len, err := d.eval(length)
	if err != nil {
		return -1, err
	}

	switch unit {
	case BitLengthUnit:
		return len, nil
	case ByteLengthUnit:
		return len * 8, nil
	}

	return -1, fmt.Errorf("unknown length unit %d", unit)
}

func (d *Decoder) String() string {
	panic("blah")
}
