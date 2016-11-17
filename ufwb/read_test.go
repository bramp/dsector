package ufwb

import (
	"bramp.net/dsector/input"
	"github.com/kylelemons/godebug/pretty"
	"io"
	"strings"
	"testing"
)

const (
	// testHeader and testFooter is prefixed to test data as it is common to all grammars
	testHeader       = `<ufwb><grammar name="Test" start="99" author="bramp@">`
	testStructHeader = testHeader + `<structure id="99">`
	testStructFooter = `</structure>` + testFooter
	testFooter       = `</grammar></ufwb>`
)

func testFile(t *testing.T, grammar *Ufwb, filename string, expectErr bool) {

	file, err := input.OpenOSFile(filename)
	if err != nil {
		t.Errorf("OpenOSFile(%q) = %q want nil error", filename, err)
		return
	}
	defer file.Close()

	fileLen, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		t.Errorf("file.Seek(0, io.SeekEnd) error %q want nil error", err)
		return
	}

	// Reset to beginning
	file.Seek(0, io.SeekStart)

	decoder := NewDecoder(grammar, file)
	got, err := decoder.Decode()

	if expectErr {
		if err == nil {
			t.Errorf("decoder.Decode(%q) = nil want error", filename)
		}
		return // We expected a error, so return
	}

	if err != nil && err != io.EOF {
		t.Errorf("decoder.Decode(%q) = %q want nil error", filename, err)
		return
	}

	if got == nil {
		t.Errorf("decoder.Decode(%q) = nil err:%q, want non-nil", filename, err)
		return
	}

	if got.Offset != 0 || got.Len != fileLen {
		t.Errorf("got {Offset: %d, Len: %d} want {Offset: %d, Len: %d}", got.Offset, got.Len, 0, fileLen)
		return
	}

	if err := got.validiate(); err != nil {
		t.Errorf("got.validiate() error = %q want nil error", err)
		return
	}

	_, err = got.Format(file)
	if err != nil {
		t.Errorf("got.Format() error = %q want nil error", err)
		return
	}
}

func TestReadNumber(t *testing.T) {
	binary := []byte("\x81\x82\x83\x84\x85\x86\x87\x88")
	var tests = []struct {
		xml      string
		wantDec  string
		wantHex  string
		wantInt  int64
		wantUint uint64
	}{
		// Unsigned
		{
			xml:      `<number id="1" type="integer" length="1" endian="big" signed="no"/>`,
			wantDec:  "129",
			wantHex:  "0x81",
			wantInt:  129,
			wantUint: 129,
		}, {
			xml:      `<number id="1" type="integer" length="2" endian="big" signed="no"/>`,
			wantDec:  "33154",
			wantHex:  "0x8182",
			wantInt:  33154,
			wantUint: 33154,
		}, {
			xml:      `<number id="1" type="integer" length="4" endian="big" signed="no"/>`,
			wantDec:  "2172814212",
			wantHex:  "0x81828384",
			wantInt:  2172814212,
			wantUint: 2172814212,
		}, {
			xml:      `<number id="1" type="integer" length="8" endian="big" signed="no"/>`,
			wantDec:  "9332165983064197000",
			wantHex:  "0x8182838485868788",
			wantInt:  -9114578090645354616,
			wantUint: 9332165983064197000,
		}, {
			xml:      `<number id="1" type="integer" length="1" endian="little" signed="no"/>`,
			wantDec:  "129",
			wantHex:  "0x81",
			wantInt:  129,
			wantUint: 129,
		}, {
			xml:      `<number id="1" type="integer" length="2" endian="little" signed="no"/>`,
			wantDec:  "33409",
			wantHex:  "0x8281",
			wantInt:  33409,
			wantUint: 33409,
		}, {
			xml:      `<number id="1" type="integer" length="4" endian="little" signed="no"/>`,
			wantDec:  "2223211137",
			wantHex:  "0x84838281",
			wantInt:  2223211137,
			wantUint: 2223211137,
		}, {
			xml:      `<number id="1" type="integer" length="8" endian="little" signed="no"/>`,
			wantDec:  "9837979819026121345",
			wantHex:  "0x8887868584838281",
			wantInt:  -0x7778797a7b7c7d7f, // The Uint value - (2^64+1)
			wantUint: 0x8887868584838281,
		},

		// Signed
		{
			xml:      `<number id="1" type="integer" length="1" endian="big" signed="yes"/>`,
			wantDec:  "-127",
			wantHex:  "0x81",
			wantInt:  -127,
			wantUint: 0xffffffffffffff81,
		}, {
			xml:      `<number id="1" type="integer" length="2" endian="big" signed="yes"/>`,
			wantDec:  "-32382",
			wantHex:  "0x8182",
			wantInt:  -32382,
			wantUint: 0xffffffffffff8182,
		}, {
			xml:      `<number id="1" type="integer" length="4" endian="big" signed="yes"/>`,
			wantDec:  "-2122153084",
			wantHex:  "0x81828384",
			wantInt:  -2122153084,
			wantUint: 0xffffffff81828384,
		}, {
			xml:      `<number id="1" type="integer" length="8" endian="big" signed="yes"/>`,
			wantDec:  "-9114578090645354616",
			wantHex:  "0x8182838485868788",
			wantInt:  -9114578090645354616,
			wantUint: 0x8182838485868788,
		}, {
			xml:      `<number id="1" type="integer" length="1" endian="little" signed="yes"/>`,
			wantDec:  "-127",
			wantHex:  "0x81",
			wantInt:  -127,
			wantUint: 0xffffffffffffff81,
		}, {
			xml:      `<number id="1" type="integer" length="2" endian="little" signed="yes"/>`,
			wantDec:  "-32127",
			wantHex:  "0x8281",
			wantInt:  -32127,
			wantUint: 0xffffffffffff8281,
		}, {
			xml:      `<number id="1" type="integer" length="4" endian="little" signed="yes"/>`,
			wantDec:  "-2071756159",
			wantHex:  "0x84838281",
			wantInt:  -2071756159,
			wantUint: 0xffffffff84838281,
		}, {
			xml:      `<number id="1" type="integer" length="8" endian="little" signed="yes"/>`,
			wantDec:  "-8608764254683430271",
			wantHex:  "0x8887868584838281",
			wantInt:  -8608764254683430271,
			wantUint: 0x8887868584838281,
		},

		// TODO Test Floats
		// TODO Test Display
		// TODO Test Bits
	}

	for _, test := range tests {
		xml := testStructHeader + test.xml + testStructFooter
		grammar, errs := ParseXmlGrammar(strings.NewReader(xml))
		if len(errs) > 0 {
			t.Errorf("ParseXmlGrammar(%q) = %q want nil error", test.xml, errs)
			continue
		}

		num, found := grammar.Get("1")
		if !found {
			t.Errorf("grammar.Get(1) = nil failed to find number element")
			continue
		}

		file := input.FromBytes(binary)
		decoder := NewDecoder(grammar, file)
		got, err := decoder.Decode()
		if err != nil {
			t.Errorf("decoder.Decode() = %q want nil error", err)
			continue
		}

		if err := got.validiate(); err != nil {
			t.Errorf("value.Validiate() = %q want nil error", err)
			continue
		}

		numValue, found := got.find(num)
		if !found {
			t.Errorf("no Number value decoded")
			continue
		}

		n := num.(*Number)
		// Decimal
		n.SetDisplay(DecDisplay)
		s, err := n.Format(file, numValue)
		if err != nil {
			t.Errorf("dec n.Format(..., %v) error = %q want nil error", numValue, err)
		}
		if s != test.wantDec {
			t.Errorf("dec n.Format(..., %v) = %q want %q", numValue, s, test.wantDec)
		}

		// Hex
		n.SetDisplay(HexDisplay)
		s, err = n.Format(file, numValue)
		if err != nil {
			t.Errorf("hex n.Format(...) error = %q want nil error", err)
		}
		if s != test.wantHex {
			t.Errorf("hex n.Format(...) = %q want %q", s, test.wantHex)
		}

		// Int64
		i, err := n.Int(file, numValue)
		if i != test.wantInt {
			t.Errorf("n.Int(...) = %d want %d", i, test.wantInt)
		}

		// Uint64
		u, err := n.Uint(file, numValue)
		if u != test.wantUint {
			t.Errorf("n.Uint(...) = %d want %d", u, test.wantUint)
		}
	}
}

func TestReadString(t *testing.T) {
	binary := []byte("abcdefghijklmnopqrstuvwxyz\x00")
	var tests = []struct {
		id         string
		xml        string // TODO Change to be a String Element
		want       Value
		wantString string
	}{
		{
			id:         "1",
			xml:        `<string id="1" type="zero-terminated"/>`,
			want:       Value{Offset: 2, Len: 25}, // 24 characters + 1 nul
			wantString: "cdefghijklmnopqrstuvwxyz",
		},
		{
			id:         "2",
			xml:        `<string id="2" type="fixed-length" length="10"/>`,
			want:       Value{Offset: 2, Len: 10},
			wantString: "cdefghijkl",
		},
		/* // TODO
		{
			xml:  `<string id="1" type="pascal"/>`,
			want: "abcdefghijklmnopqrstuvwxyz",
		},
		*/
	}

	for _, test := range tests {
		xml := testStructHeader + test.xml + testStructFooter
		grammar, errs := ParseXmlGrammar(strings.NewReader(xml))
		if len(errs) > 0 {
			t.Errorf("ParseXmlGrammar(%q) = %q want nil error", test.xml, errs)
			continue
		}

		str, found := grammar.Get(test.id)
		if !found {
			t.Errorf("grammar.Get(%q) = nil failed to find string element", test.id)
			continue
		}

		file := input.FromBytes(binary)
		file.Seek(2, io.SeekStart) // Seek two bytes to check for offset assumption

		decoder := NewDecoder(grammar, file)
		got, err := decoder.Decode()
		if err != nil {
			t.Errorf("decoder.Decode() error = %q want nil error", err)
			continue
		}

		if err := got.validiate(); err != nil {
			t.Errorf("value.Validiate() = %q want nil error", err)
			continue
		}

		strValue, found := got.find(str)
		if !found {
			t.Errorf("no String value decoded")
			continue
		}

		if strValue.Offset != test.want.Offset || strValue.Len != test.want.Len {
			t.Errorf("strValue{Offset: %d, Len: %d} != want{Offset: %d, Len: %d}",
				strValue.Offset, strValue.Len, test.want.Offset, test.want.Len)
			continue
		}

		s, err := str.Format(file, strValue)
		if err != nil {
			t.Errorf("str.Format(...) error = %q want nil error", err)
		}
		if s != test.wantString {
			t.Errorf("str.Format(...) = %q want %q", s, test.wantString)
		}
	}
}

func TestBoundReads(t *testing.T) {
	binary := []byte("abcdefghijklmnopqrstuvwxyz\x00")
	var tests = []struct {
		element Element
	}{
		{
			element: &String{typ: "fixed-length", length: ConstExpression(2)},
		}, {
			element: &String{typ: "fixed-length", length: ConstExpression(10)},
		}, /* TODO {
			element: &String{typ: "pascal"},
		},*/{
			element: &Number{length: ConstExpression(2)},
		}, {
			element: &Number{length: ConstExpression(8)},
		}, {
			element: &Binary{length: ConstExpression(2)},
		}, {
			element: &Binary{length: ConstExpression(8)},
		},
	}

	for _, test := range tests {
		file := input.FromBytes(binary)

		// 1 byte bound
		decoder := NewDecoderWithBounds(nil, file, 0, 1)

		got, err := test.element.Read(decoder)

		// TODO Change this to print out the element's XML form (to make it easier to read test output)
		if got != nil {
			t.Errorf("[%s] Read(..) = %q, want nil", test.element, got)
		}

		if err == nil {
			t.Errorf("[%s] Read(..) err = nil, want io.ErrUnexpectedEOF", test.element)
		}
	}
}

func TestEOFReads(t *testing.T) {
	empty := []byte{}
	var tests = []struct {
		element Element
	}{
		{
			element: &String{typ: "zero-terminated"},
		}, {
			element: &String{typ: "delimiter-terminated", delimiter: ','},
		}, {
			element: &String{typ: "fixed-length", length: ConstExpression(1)},
		}, {
			element: &String{typ: "fixed-length", length: ConstExpression(10)},
		}, {
			element: &String{typ: "pascal"},
		}, {
			element: &Number{length: ConstExpression(1)},
		}, {
			element: &Number{length: ConstExpression(8)},
		}, {
			element: &Binary{length: ConstExpression(1)},
		}, {
			element: &Binary{length: ConstExpression(8)},
		},
	}

	for _, test := range tests {
		file := input.FromBytes(empty)
		decoder := NewDecoder(nil, file)

		got, err := test.element.Read(decoder)

		// TODO Change this to print out the element's XML form (to make it easier to read test output)
		if got != nil {
			t.Errorf("[%s] Read(..) = %q, want nil", test.element, got)
		}

		if err != io.EOF {
			t.Errorf("[%s] Read(..) err = %q, want io.EOF", test.element, err)
		}
	}
}

func TestShortReads(t *testing.T) {
	short := []byte{1}
	var tests = []struct {
		element Element
	}{
		{
			element: &String{typ: "zero-terminated"},
		}, {
			element: &String{typ: "delimiter-terminated", delimiter: ','},
		}, {
			element: &String{typ: "fixed-length", length: ConstExpression(2)},
		}, {
			element: &String{typ: "fixed-length", length: ConstExpression(10)},
		}, {
			element: &String{typ: "pascal"},
		}, {
			element: &Number{length: ConstExpression(2)},
		}, {
			element: &Number{length: ConstExpression(8)},
		}, {
			element: &Binary{length: ConstExpression(2)},
		}, {
			element: &Binary{length: ConstExpression(8)},
		},
	}

	for _, test := range tests {
		file := input.FromBytes(short)
		decoder := NewDecoder(nil, file)

		got, err := test.element.Read(decoder)

		// TODO Change this to print out the element's XML form (to make it easier to read test output)
		if got != nil {
			t.Errorf("[%s] Read(..) = %q, want nil", test.element, got)
		}

		if err == nil {
			t.Errorf("[%s] Read(..) err = nil, want non-nil", test.element)
		}
	}
}

func TestRepeating(t *testing.T) {
	binary := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8}
	xml := testHeader +
		`<structure name="Colours" id="99">
			<structure name="Colour" id="2" repeatmax="unlimited">
				<number name="Red"   id="3" type="integer" length="1"/>
				<number name="Green" id="4" type="integer" length="1"/>
				<number name="Blue"  id="5" type="integer" length="1"/>
			</structure>
		</structure>` + testFooter

	want := `Test: (1 children)
  [0] Colours: (3 children)
    [0] Colour: (3 children)
      [0] Red: 0
      [1] Green: 1
      [2] Blue: 2
    [1] Colour: (3 children)
      [0] Red: 3
      [1] Green: 4
      [2] Blue: 5
    [2] Colour: (3 children)
      [0] Red: 6
      [1] Green: 7
      [2] Blue: 8`

	grammar, errs := ParseXmlGrammar(strings.NewReader(xml))
	if len(errs) > 0 {
		t.Errorf("ParseXmlGrammar(%q) = %q want nil error", xml, errs)
		return
	}

	file := input.FromBytes(binary)
	decoder := NewDecoder(grammar, file)

	value, err := decoder.Decode()
	if err != nil {
		t.Errorf("decoder.Decode() error = %q want nil error", err)
		return
	}

	if err := value.validiate(); err != nil {
		t.Errorf("value.Validiate() = %q want nil error", err)
		return
	}

	got, err := grammar.Format(file, value)
	if err != nil {
		t.Errorf("grammar.Format(...) error = %q want nil error", err)
	}

	if diff := pretty.Compare(strings.TrimSpace(got), want); diff != "" {
		t.Errorf("grammar.Format(...) = -got +want:\n%s", diff)
	}
}
