package ufwb

import (
	"fmt"
	"strings"
	"encoding/hex"

)

func leftPad(s string, pad string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(pad, width - len(s)) + s
}

func (g *Grammar) Format(f File, value *Value) (string, error) {
	return g.Start.Format(f, value)
}

func (n *Structure) Format(f File, value *Value) (string, error) {
	return "TODO", nil
}

func (n *String) Format(f File, value *Value) (string, error) {
	//return fmt.Sprintf("%q", )
	return "TODO", nil
}

// format returns a formatted string of the given int. The int must be one of int{8,16,32,64} or
// uint{8,16,32,64} types.
func (n *Number) format(i interface{}) (string, error) {
	base := n.Display().Base()
	if base < 2 || base > 36 {
		return "",  &validationError{e: n, msg: fmt.Sprintf("invalid base %d", base)}
	}

	return formatInt(i, base, n.Bits())
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

func (n *Number) Format(f File, value *Value) (string, error) {
	i, err := n.int(f, value)
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

func (b *Binary) Format(f File, value *Value) (string, error) {
	bs, err := b.Bytes(f, value)
	if err != nil {
		return "", err
	}

	return b.format(bs)
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
