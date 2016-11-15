// TODO Change Format API to have a FormatTo(.., io.writer), or similar, so we can avoid all the string concat.
package ufwb

import (
	"encoding/hex"
	"fmt"
	"strings"

	"bytes"
	"io"
	"reflect"
)

var (
	indent = 0 // Bad global var for format identing
)

func leftPad(s string, pad string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(pad, width-len(s)) + s
}

func (u *Ufwb) Format(file io.ReaderAt, value *Value) (string, error) {
	indent = 0
	return u.Grammar.Format(file, value)
}

func (p *Padding) Format(file io.ReaderAt, value *Value) (string, error) {
	return fmt.Sprintf("<padding len:%d>", value.Len), nil
}

func (g *Grammar) Format(file io.ReaderAt, value *Value) (string, error) {
	var buffer bytes.Buffer

	buffer.WriteString(g.Name())
	buffer.WriteString(": ")

	str, err := g.Start.Format(file, value)
	if err != nil {
		return "<format error>", err
	}

	buffer.WriteString(str)

	return buffer.String(), nil
}

func (n *Structure) Format(file io.ReaderAt, value *Value) (string, error) {

	indent++

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("(%d children)", len(value.Children)))
	buffer.WriteString("\n")

	pad := ""
	for i := 0; i < indent; i++ {
		pad += "  "
	}

	for i, child := range value.Children {
		buffer.WriteString(pad)
		buffer.WriteString(fmt.Sprintf("[%d] ", i))
		buffer.WriteString(child.Name())
		buffer.WriteString(": ")
		child.Format(file)

		value, err := child.Format(file)
		if err != nil {
			indent--
			return "<format error>", err
		}
		buffer.WriteString(strings.TrimSpace(value))
		buffer.WriteString("\n")
	}

	indent--

	return buffer.String(), nil
}

func (s *String) Format(file io.ReaderAt, value *Value) (string, error) {

	b := make([]byte, value.Len, value.Len)
	n, err := file.ReadAt(b, value.Offset)
	if err != nil {
		return string(b[:n]), &validationError{e: s, err: err}
	}

	switch s.Typ() {
	case "zero-terminated":
		// Strip the nul character if it exists
		if b[n-1] == 0 {
			return string(b[:n-1]), nil
		}

	case "pascal":
		// Skip the length byte at the beginning of the string
		panic("TODO pascal string format")
	}

	return string(b), nil
}

// format returns a formatted string of the given int. The int must be one of int{8,16,32,64} or
// uint{8,16,32,64} types.
func (n *Number) format(i interface{}) (string, error) {
	base := n.Display().Base()
	if base < 2 || base > 36 {
		return "", &validationError{e: n, err: fmt.Errorf("invalid base %d", base)}
	}

	// TODO this needs fixing for non-multiple of 8 bit numbers
	bits := int(reflect.ValueOf(i).Type().Size()) * 8

	return formatInt(i, base, bits)
}

func (n *Number) formatValues() ([]string, error) {
	var ret []string
	for _, v := range n.Values() {
		s, err := n.format(v.value)
		if err != nil {
			return nil, err
		}
		ret = append(ret, s)
	}
	return ret, nil
}

func (n *Number) Format(file io.ReaderAt, value *Value) (string, error) {
	i, err := n.int(file, value)
	if err != nil {
		return "", err
	}

	s, err := n.format(i)
	if err == nil {
		if fv, ok := value.Extra.(*FixedValue); ok && fv.name != "" {
			s += fmt.Sprintf(" (%s)", fv.name)
		}
	}
	return s, err
}

func (b *Binary) format(bs []byte) (string, error) {
	// TODO Maybe use b.Length() to change the output?
	s := ""
	if len(bs) <= 8 {
		s = hex.EncodeToString(bs)
	} else {
		s = hex.EncodeToString(bs[:6]) + "..."
	}
	return s, nil
}

func (b *Binary) formatValues() ([]string, error) {
	var ret []string
	for _, v := range b.Values() {
		s, err := b.format(v.value)
		if err != nil {
			return nil, err
		}
		ret = append(ret, s)
	}
	return ret, nil
}

func (b *Binary) Format(file io.ReaderAt, value *Value) (string, error) {
	bs, err := b.Bytes(file, value)
	if err != nil {
		return "", err
	}

	s, err := b.format(bs)
	if err == nil {
		if fv, ok := value.Extra.(*FixedBinaryValue); ok {
			s += fmt.Sprintf(" (%s)", fv.name)
		} else {
			s += fmt.Sprintf(" (%d bytes)", len(bs))
		}
	}

	return s, err
}

func (n *Custom) Format(file io.ReaderAt, value *Value) (string, error) {
	panic("TODO")
}

func (n *GrammarRef) Format(file io.ReaderAt, value *Value) (string, error) {
	panic("TODO")
}

func (n *Offset) Format(file io.ReaderAt, value *Value) (string, error) {
	panic("TODO")
}

func (n *Script) Format(file io.ReaderAt, value *Value) (string, error) {
	return "<script ran>", nil
}

func (n *StructRef) Format(file io.ReaderAt, value *Value) (string, error) {
	return n.Structure().Format(file, value)
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
