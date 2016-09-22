// TODO Move this file into a ints package
package ufwb

import (
	"reflect"
	"fmt"
	"strconv"
	"encoding/binary"
	"io"
)

// intEqual compares two integers stored in interfaces. Returns true if equal (regradless of bitsize)
func intEqual(a, b interface{}) bool {
	// TODO improve the comparasion to avoid going via strings
	af, err := formatInt(a, 16, 64)
	if err != nil {
		panic(err) // TODO Don't panic
	}
	bf, err := formatInt(b, 16, 64)
	if err != nil {
		panic(err)
	}
	return af == bf
}

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

// readInt returns the integer stored in f. The returned
// integer is one of int{8,16,32,64} or uint{8,16,32,64} depending
// on the width and sign of the integer.
func readInt(r io.Reader, len int64, signed bool, order binary.ByteOrder) (interface{}, error) {

	// Create a correctly sized int. This is so binary.Read reads the correct length
	// TODO Rewrite this so we can handle say 24 bit ints
	var i interface{}
	if signed {
		switch len {
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
		switch len {
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
		return 0, fmt.Errorf("unsupported number length: %d",len)
	}

	err := binary.Read(r, order, i)
	// Strip the pointer from the interface
	// TODO Consider refactoring, so ints are always either int64 or uint64
	return reflect.ValueOf(i).Elem().Interface(), err
}


func formatIntPad(s string, base int, bitSize int) string {
	// TODO Base 2
	if base == 16 {
		return "0x" + leftPad(s, "0", bitSize / 4)
	}

	return s
}

func formatInt(i interface{}, base int, bits int) (string, error) {

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
