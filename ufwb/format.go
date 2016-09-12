package ufwb

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"encoding/hex"
	log "github.com/Sirupsen/logrus"

)

func parseInt(s string, base int, bitSize int, signed bool) (interface{}, error) {
	n, err := strconv.ParseUint(s, base, bitSize)

	if signed && err == nil {
		cutoff := uint64(1 << uint(bitSize-1))
		// ParseInt doesn't handle signed hex numbers, so we do it ourselves
		if n >= cutoff {
			return int64(n - cutoff) - int64(cutoff), nil
		}
		return int64(n), nil
	}
	return n, err
}

func leftPad(s string, pad string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(pad, width - len(s)) + s
}

func formatIntPad(s string, base int, bitSize int) string {
	// TODO Base 2
	if base == 16 {
		return "0x" + leftPad(s, "0", bitSize / 4)
	}

	return s
}

func formatInt(i interface{}, base int, bits int) (string, error) {

	log.Debugf("%v %v %v", i, base, bits)

	switch i.(type) {
	// TODO Consider refactoring, so ints are always either int64 or uint64
	case int8, int16, int32, int64:
		n := reflect.ValueOf(i).Int()

		// FormatInt will print negative hex numbers with a minus sign infront
		// Instead we flip the sign and print as unsigned.
		if base == 16 && n < 0 {
			cutoff := int64(1 << uint(bits-1))
			u := uint64(n + cutoff) + uint64(cutoff)

			// Mask off the high bits (which will all be one now)
			mask := uint64(1 << uint(bits)) - 1
			return formatInt(u & mask, base, bits)
		}

		return formatIntPad(strconv.FormatInt(n, base), base, bits), nil

	case uint8, uint16, uint32, uint64:
		n := reflect.ValueOf(i).Uint()
		return formatIntPad(strconv.FormatUint(n, base), base, bits), nil
	}

	panic(fmt.Sprintf("unknown integer type %T", i))
}


func (g *Grammar) Format(f File, value *Value) (string, error) {
	return g.Start.Format(f, value)
}

func (n *Structure) Format(f File, value *Value) (string, error) {
	panic("TODO")
}

func (n *String) Format(f File, value *Value) (string, error) {
	panic("TODO")
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
