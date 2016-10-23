package ufwb

import (
	"github.com/kylelemons/godebug/pretty"
	"io"
	"strings"
	"testing"
	"bramp.net/dsector/input"
)

const (
	samplesPath = "../samples"

	// commonHeader and commonFooter is prefixed to test data as it is common to all grammars
	commonHeader = `<ufwb><grammar name="Test" start="1" author="bramp@"><structure id="99">`
	commonFooter = `</structure></grammar></ufwb>`
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
		xml     string
		wantDec string
		wantHex string
	}{
		// Unsigned
		{
			xml:     `<number id="1" type="integer" length="1" endian="big" signed="no"/>`,
			wantDec: "129",
			wantHex: "0x81",
		}, {
			xml:     `<number id="1" type="integer" length="2" endian="big" signed="no"/>`,
			wantDec: "33154",
			wantHex: "0x8182",
		}, {
			xml:     `<number id="1" type="integer" length="4" endian="big" signed="no"/>`,
			wantDec: "2172814212",
			wantHex: "0x81828384",
		}, {
			xml:     `<number id="1" type="integer" length="8" endian="big" signed="no"/>`,
			wantDec: "9332165983064197000",
			wantHex: "0x8182838485868788",
		}, {
			xml:     `<number id="1" type="integer" length="1" endian="little" signed="no"/>`,
			wantDec: "129",
			wantHex: "0x81",
		}, {
			xml:     `<number id="1" type="integer" length="2" endian="little" signed="no"/>`,
			wantDec: "33409",
			wantHex: "0x8281",
		}, {
			xml:     `<number id="1" type="integer" length="4" endian="little" signed="no"/>`,
			wantDec: "2223211137",
			wantHex: "0x84838281",
		}, {
			xml:     `<number id="1" type="integer" length="8" endian="little" signed="no"/>`,
			wantDec: "9837979819026121345",
			wantHex: "0x8887868584838281",
		},

		// Signed
		{
			xml:     `<number id="1" type="integer" length="1" endian="big" signed="yes"/>`,
			wantDec: "-127",
			wantHex: "0x81",
		}, {
			xml:     `<number id="1" type="integer" length="2" endian="big" signed="yes"/>`,
			wantDec: "-32382",
			wantHex: "0x8182",
		}, {
			xml:     `<number id="1" type="integer" length="4" endian="big" signed="yes"/>`,
			wantDec: "-2122153084",
			wantHex: "0x81828384",
		}, {
			xml:     `<number id="1" type="integer" length="8" endian="big" signed="yes"/>`,
			wantDec: "-9114578090645354616",
			wantHex: "0x8182838485868788",
		}, {
			xml:     `<number id="1" type="integer" length="1" endian="little" signed="yes"/>`,
			wantDec: "-127",
			wantHex: "0x81",
		}, {
			xml:     `<number id="1" type="integer" length="2" endian="little" signed="yes"/>`,
			wantDec: "-32127",
			wantHex: "0x8281",
		}, {
			xml:     `<number id="1" type="integer" length="4" endian="little" signed="yes"/>`,
			wantDec: "-2071756159",
			wantHex: "0x84838281",
		}, {
			xml:     `<number id="1" type="integer" length="8" endian="little" signed="yes"/>`,
			wantDec: "-8608764254683430271",
			wantHex: "0x8887868584838281",
		},

		// TODO Test Floats
		// TODO Test Display
		// TODO Test Bits
	}

	for _, test := range tests {
		xml := commonHeader + test.xml + commonFooter
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
		}

		// Decimal
		num.(*Number).SetDisplay(DecDisplay)
		s, err := num.Format(file, got)
		if err != nil {
			t.Errorf("dec num.Format(...) error = %q want nil error", err)
		}
		if s != test.wantDec {
			t.Errorf("dec num.Format(...) = %q want %q", s, test.wantDec)
		}

		// Hex
		num.(*Number).SetDisplay(HexDisplay)
		s, err = num.Format(file, got)
		if err != nil {
			t.Errorf("hex num.Format(...) error = %q want nil error", err)
		}
		if s != test.wantHex {
			t.Errorf("hex num.Format(...) = %q want %q", s, test.wantHex)
		}
	}
}

func TestReadString(t *testing.T) {
	binary := []byte("abcdefghijklmnopqrstuvwxyz\x00")
	var tests = []struct {
		xml        string // TODO Change to be a String Element
		want       Value
		wantString string
	}{
		{
			xml:        `<string id="1" type="zero-terminated"/>`,
			want:       Value{Offset: 2, Len: 25}, // 24 characters + 1 nul
			wantString: "cdefghijklmnopqrstuvwxyz",
		},
		{
			xml:        `<string id="1" type="fixed-length" length="10"/>`,
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
		xml := commonHeader + test.xml + commonFooter
		grammar, errs := ParseXmlGrammar(strings.NewReader(xml))
		if len(errs) > 0 {
			t.Errorf("ParseXmlGrammar(%q) = %q want nil error", test.xml, errs)
			continue
		}

		str, found := grammar.Get("1")
		if !found {
			t.Errorf("grammar.Get(1) = nil failed to find string element")
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

		if got.Offset != test.want.Offset || got.Len != test.want.Len {
			t.Errorf("got{Offset: %d, Len: %d} != want{Offset: %d, Len: %d}",
				got.Offset, got.Len, test.want.Offset, test.want.Len)
			continue
		}

		s, err := str.Format(file, got)
		if err != nil {
			t.Errorf("str.Format(...) error = %q want nil error", err)
		}
		if s != test.wantString {
			t.Errorf("str.Format(...) = %q want %q", s, test.wantString)
		}
	}
}

func TestRepeating(t *testing.T) {
	binary := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8}
	xml := commonHeader +
		`<structure name="Colours" id="1">
				<structure name="Colour" id="2" repeatmax="unlimited">
					<number name="Red"   id="3" type="integer" length="1"/>
					<number name="Green" id="4" type="integer" length="1"/>
					<number name="Blue"  id="5" type="integer" length="1"/>
				</structure>
			</structure>` +
		commonFooter

	want :=
		`Colours: (3 children)
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
