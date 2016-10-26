package ufwb

import "testing"

var intTests = []struct {
	in     string
	base   int
	bits   int
	signed bool
	out    interface{}
}{
	// Base 10
	// Signed
	{in: "0", base: 10, bits: 32, signed: true, out: int64(0)},
	{in: "1", base: 10, bits: 32, signed: true, out: int64(1)},
	{in: "2147483647", base: 10, bits: 32, signed: true, out: int64(2147483647)},
	{in: "-2147483648", base: 10, bits: 32, signed: true, out: int64(-2147483648)},
	{in: "-19088744", base: 10, bits: 32, signed: true, out: int64(-19088744)},
	{in: "-1", base: 10, bits: 32, signed: true, out: int64(-1)},

	{in: "0", base: 10, bits: 64, signed: true, out: int64(0)},
	{in: "9223372036854775807", base: 10, bits: 64, signed: true, out: int64(9223372036854775807)},
	{in: "-9223372036854775808", base: 10, bits: 64, signed: true, out: int64(-9223372036854775808)},
	{in: "-1", base: 10, bits: 64, signed: true, out: int64(-1)},

	// Unsigned
	{in: "0", base: 10, bits: 32, signed: false, out: uint64(0)},
	{in: "1", base: 10, bits: 32, signed: false, out: uint64(1)},
	{in: "2147483647", base: 10, bits: 32, signed: false, out: uint64(2147483647)},
	{in: "2147483648", base: 10, bits: 32, signed: false, out: uint64(2147483648)},
	{in: "4275878552", base: 10, bits: 32, signed: false, out: uint64(4275878552)},
	{in: "4294967295", base: 10, bits: 32, signed: false, out: uint64(4294967295)},

	{in: "0", base: 10, bits: 64, signed: false, out: uint64(0)},
	{in: "9223372036854775807", base: 10, bits: 64, signed: false, out: uint64(9223372036854775807)},
	{in: "9223372036854775808", base: 10, bits: 64, signed: false, out: uint64(9223372036854775808)},
	{in: "18446744073709551615", base: 10, bits: 64, signed: false, out: uint64(18446744073709551615)},

	// Base 16
	// Signed
	{in: "0x00000000", base: 16, bits: 32, signed: true, out: int64(0)},
	{in: "0x00000001", base: 16, bits: 32, signed: true, out: int64(1)},
	{in: "0x7fffffff", base: 16, bits: 32, signed: true, out: int64(2147483647)},
	{in: "0x80000000", base: 16, bits: 32, signed: true, out: int64(-2147483648)},
	{in: "0xfedcba98", base: 16, bits: 32, signed: true, out: int64(-19088744)},
	{in: "0xffffffff", base: 16, bits: 32, signed: true, out: int64(-1)},

	{in: "0x0000000000000000", base: 16, bits: 64, signed: true, out: int64(0)},
	{in: "0x7fffffffffffffff", base: 16, bits: 64, signed: true, out: int64(9223372036854775807)},
	{in: "0x8000000000000000", base: 16, bits: 64, signed: true, out: int64(-9223372036854775808)},
	{in: "0xffffffffffffffff", base: 16, bits: 64, signed: true, out: int64(-1)},

	// Unsigned
	{in: "0x00000000", base: 16, bits: 32, signed: false, out: uint64(0)},
	{in: "0x00000001", base: 16, bits: 32, signed: false, out: uint64(1)},
	{in: "0x7fffffff", base: 16, bits: 32, signed: false, out: uint64(2147483647)},
	{in: "0x80000000", base: 16, bits: 32, signed: false, out: uint64(2147483648)},
	{in: "0xfedcba98", base: 16, bits: 32, signed: false, out: uint64(4275878552)},
	{in: "0xffffffff", base: 16, bits: 32, signed: false, out: uint64(4294967295)},

	{in: "0x0000000000000000", base: 16, bits: 64, signed: false, out: uint64(0)},
	{in: "0x7fffffffffffffff", base: 16, bits: 64, signed: false, out: uint64(9223372036854775807)},
	{in: "0x8000000000000000", base: 16, bits: 64, signed: false, out: uint64(9223372036854775808)},
	{in: "0xffffffffffffffff", base: 16, bits: 64, signed: false, out: uint64(18446744073709551615)},
}

func TestParseInt(t *testing.T) {

	for _, test := range intTests {
		got, err := parseInt(test.in, 0, test.bits, test.signed)
		if err != nil {
			t.Errorf("parseInt(%q, %d, %t) error = %q want nil", test.in, test.bits, test.signed, err)
			continue
		}
		if got != test.out {
			t.Errorf("parseInt(%q, %d, %t) = %d want %d", test.in, test.bits, test.signed, got, test.out)
			continue
		}
	}
}

func TestFormatInt(t *testing.T) {
	for _, test := range intTests {
		got, err := formatInt(test.out, test.base, test.bits)
		if err != nil {
			t.Errorf("formatInt(%v, %d, %d) error = %q want nil", test.out, 16, test.bits, err)
			continue
		}
		if got != test.in {
			t.Errorf("formatInt(%v, %d, %d) = %q want %q", test.out, 16, test.bits, got, test.in)
			continue
		}
	}
}
