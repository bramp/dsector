// TODO Change Format API to have a FormatTo(.., io.writer), or similar, so we can avoid all the string concat.
package ufwb

import (
	"fmt"
	"strings"
	"encoding/hex"

	"reflect"
	"bytes"
	"io"
)

var (
	indent = 0 // Bad global var for format identing
)

func leftPad(s string, pad string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(pad, width - len(s)) + s
}

func (u *Ufwb) Format(file io.ReaderAt, value *Value) (string, error) {
	indent = 0
	return u.Grammar.Format(file, value)
}

func (p *Padding) Format(file io.ReaderAt, value *Value) (string, error) {
	return fmt.Sprint("<padding len:%d>", value.Len), nil
}

func (g *Grammar) Format(file io.ReaderAt, value *Value) (string, error) {
	var buffer bytes.Buffer

	buffer.WriteString(g.Start.Name());
	buffer.WriteString(": ")

	str, err := g.Start.Format(file, value)
	if err != nil {
		return "TODO ERR", err
	}

	buffer.WriteString(str)

	return buffer.String(), nil
}

func (n *Structure) Format(file io.ReaderAt, value *Value) (string, error) {

	indent++

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("(%d children)", len(value.Children)));
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
			return "TODO ERR", err
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
		if b[n - 1] == 0 {
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
		return "",  &validationError{e: n, err: fmt.Errorf("invalid base %d", base)}
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
	return n.format(i)
}


func (b *Binary) format(bs []byte) (string, error) {
	// TODO Maybe use b.Length() to change the output?
	if len(bs) > 8 {
		return hex.EncodeToString(bs[:6]) + "..", nil
	}
	return hex.EncodeToString(bs), nil
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

	return b.format(bs)
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

func (n *ScriptElement) Format(file io.ReaderAt, value *Value) (string, error) {
	panic("TODO")
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
