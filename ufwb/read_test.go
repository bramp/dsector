package ufwb

import (
	"bytes"
	"io/ioutil"
	"path"
	"strings"
	"testing"
	"io"
	log "github.com/Sirupsen/logrus"
)

const (
	samplesPath = "../samples"

	// commonHeader and commonFooter is prefixed to test data as it is common to all grammars
	commonHeader = `<ufwb><grammar name="Test" start="1" author="bramp@"><structure id="99">`
	commonFooter = `</structure></grammar></ufwb>`
)

func init() {
	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

func TestParserPng(t *testing.T) {

	lang := "png"
	langFile := lang + ".grammar"

	in, err := readGrammar(langFile)
	if err != nil {
		t.Fatalf("readGrammar(%q) = %s want nil error", langFile, err)
	}

	// Parse
	grammar, errs := ParseXmlGrammar(bytes.NewReader(in))
	if len(errs) > 0 {
		t.Fatalf("Parse(%q) = %q want nil error", langFile, errs)
	}

	// Now read each sample png file:
	root := path.Join(samplesPath, lang)
	files, err := ioutil.ReadDir(root)
	if err != nil {
		t.Fatalf("ioutil.ReadDir(%q) = %q want nil error", root, err)
	}

	for _, sample := range files {
		filename := path.Join(root, sample.Name())
		if !strings.HasSuffix(filename, ".png") {
			continue
		}

		file, err := OpenOSFile(filename)
		if err != nil {
			log.Debug(err)
			t.Errorf("OpenOSFile(%q) = %q want nil error", filename, err)
			continue
		}

		decoder := NewDecoder(grammar, file)
		got, err := decoder.Decode()
		if err != nil && err != io.EOF {
			log.Debug(err)
			t.Errorf("decoder.Decode(%q) = %q want nil error", sample.Name(), err)
			file.Close()
			continue
		}

		if err := got.validiate(); err != nil {
			log.Debug(err)
			t.Errorf("value.Validiate() = %q want nil error", err)
			return
		}

		s, _ := got.Format(file)
		t.Logf("%s\n%s", filename, s)
		file.Close()
	}
}

func TestParserNumber(t *testing.T) {
	binary := []byte("\x81\x82\x83\x84\x85\x86\x87\x88")
	var tests = []struct {
		xml     string
		wantDec string
		wantHex string
	}{
		// Unsigned
		{
			xml:  `<number id="1" type="integer" length="1" endian="big" signed="no"/>`,
			wantDec: "129",
			wantHex: "0x81",
		}, {
			xml:  `<number id="1" type="integer" length="2" endian="big" signed="no"/>`,
			wantDec: "33154",
			wantHex: "0x8182",
		}, {
			xml:  `<number id="1" type="integer" length="4" endian="big" signed="no"/>`,
			wantDec: "2172814212",
			wantHex: "0x81828384",
		}, {
			xml:  `<number id="1" type="integer" length="8" endian="big" signed="no"/>`,
			wantDec: "9332165983064197000",
			wantHex: "0x8182838485868788",
		}, {
			xml:  `<number id="1" type="integer" length="1" endian="little" signed="no"/>`,
			wantDec: "129",
			wantHex: "0x81",
		}, {
			xml:  `<number id="1" type="integer" length="2" endian="little" signed="no"/>`,
			wantDec: "33409",
			wantHex: "0x8281",
		}, {
			xml:  `<number id="1" type="integer" length="4" endian="little" signed="no"/>`,
			wantDec: "2223211137",
			wantHex: "0x84838281",
		}, {
			xml:  `<number id="1" type="integer" length="8" endian="little" signed="no"/>`,
			wantDec: "9837979819026121345",
			wantHex: "0x8887868584838281",
		},

		// Signed
		{
			xml:  `<number id="1" type="integer" length="1" endian="big" signed="yes"/>`,
			wantDec: "-127",
			wantHex: "0x81",
		}, {
			xml:  `<number id="1" type="integer" length="2" endian="big" signed="yes"/>`,
			wantDec: "-32382",
			wantHex: "0x8182",
		}, {
			xml:  `<number id="1" type="integer" length="4" endian="big" signed="yes"/>`,
			wantDec: "-2122153084",
			wantHex: "0x81828384",
		}, {
			xml:  `<number id="1" type="integer" length="8" endian="big" signed="yes"/>`,
			wantDec: "-9114578090645354616",
			wantHex: "0x8182838485868788",
		}, {
			xml:  `<number id="1" type="integer" length="1" endian="little" signed="yes"/>`,
			wantDec: "-127",
			wantHex: "0x81",

		}, {
			xml:  `<number id="1" type="integer" length="2" endian="little" signed="yes"/>`,
			wantDec: "-32127",
			wantHex: "0x8281",
		}, {
			xml:  `<number id="1" type="integer" length="4" endian="little" signed="yes"/>`,
			wantDec: "-2071756159",
			wantHex: "0x84838281",
		}, {
			xml:  `<number id="1" type="integer" length="8" endian="little" signed="yes"/>`,
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
			t.Errorf("ParseGrammar(%q) = %q want nil error", test.xml, errs)
			continue
		}

		num, found := grammar.Get("1")
		if !found {
			t.Errorf("grammar.Get(1) = nil failed to find number element")
			continue
		}

		file := NewFileFromBytes(binary)
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


		//if diff := pretty.Compare(got, test.want); diff != "" {
		//	t.Errorf("decoder.Decode() = -got +want:\n%s", diff)
		//}
	}
}
