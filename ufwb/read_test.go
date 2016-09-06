package ufwb_test

import (
	"bramp.net/dsector/ufwb"
	"bytes"
	"io/ioutil"
	"path"
	"strings"
	"testing"
)

const (
	samplesPath = "../samples"

	// commonHeader and commonFooter is prefixed to test data as it is common to all grammars
	commonHeader = `<ufwb><grammar name="Test" start="1" author="bramp@"><structure id="99">`
	commonFooter = `</structure></grammar></ufwb>`
)

func TestParserPng(t *testing.T) {

	lang := "png"
	langFile := lang + ".grammar"

	in, err := readGrammar(langFile)
	if err != nil {
		t.Fatalf("readGrammar(%q) = %s want nil error", langFile, err)
	}

	// Parse
	grammar, errs := ufwb.ParseXmlGrammar(bytes.NewReader(in))
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

		file, err := ufwb.OpenOSFile(filename)
		if err != nil {
			t.Errorf("ufwb.OpenOSFile(%q) = %q want nil error", filename, err)
			continue
		}

		decoder := ufwb.NewDecoder(grammar, file)
		got, err := decoder.Decode()
		if err != nil {
			t.Errorf("decoder.Decode() = %q want nil error", err)
			file.Close()
			continue
		}

		if err := got.Validiate(); err != nil {
			t.Errorf("value.Validiate() = %q want nil error", err)
		}

		s, _ := got.Format(file)
		t.Logf("%s\n%s", filename, s)
		file.Close()
	}
}

func TestParserNumber(t *testing.T) {
	binary := []byte("\x81\x82\x83\x84\x85\x86\x87\x88")
	var tests = []struct {
		xml  string
		want string
	}{
		// Unsigned
		{
			xml:  `<number id="1" type="integer" length="1" endian="big" signed="no"/>`,
			want: "129",
		}, {
			xml:  `<number id="1" type="integer" length="2" endian="big" signed="no"/>`,
			want: "33154",
		}, {
			xml:  `<number id="1" type="integer" length="4" endian="big" signed="no"/>`,
			want: "2172814212",
		}, {
			xml:  `<number id="1" type="integer" length="8" endian="big" signed="no"/>`,
			want: "9332165983064197000",
		}, {
			xml:  `<number id="1" type="integer" length="1" endian="little" signed="no"/>`,
			want: "129",
		}, {
			xml:  `<number id="1" type="integer" length="2" endian="little" signed="no"/>`,
			want: "33409",
		}, {
			xml:  `<number id="1" type="integer" length="4" endian="little" signed="no"/>`,
			want: "2223211137",
		}, {
			xml:  `<number id="1" type="integer" length="8" endian="little" signed="no"/>`,
			want: "9837979819026121345",
		},

		// Signed
		{
			xml:  `<number id="1" type="integer" length="1" endian="big" signed="yes"/>`,
			want: "-127",
		}, {
			xml:  `<number id="1" type="integer" length="2" endian="big" signed="yes"/>`,
			want: "-32382",
		}, {
			xml:  `<number id="1" type="integer" length="4" endian="big" signed="yes"/>`,
			want: "-2122153084",
		}, {
			xml:  `<number id="1" type="integer" length="8" endian="big" signed="yes"/>`,
			want: "-9114578090645354616",
		}, {
			xml:  `<number id="1" type="integer" length="1" endian="little" signed="yes"/>`,
			want: "-127",
		}, {
			xml:  `<number id="1" type="integer" length="2" endian="little" signed="yes"/>`,
			want: "-32127",
		}, {
			xml:  `<number id="1" type="integer" length="4" endian="little" signed="yes"/>`,
			want: "-2071756159",
		}, {
			xml:  `<number id="1" type="integer" length="8" endian="little" signed="yes"/>`,
			want: "-8608764254683430271",
		},

		// TODO Test Floats
		// TODO Test Display
		// TODO Test Bits
	}

	for _, test := range tests {
		xml := commonHeader + test.xml + commonFooter
		g, errs := ufwb.ParseXmlGrammar(strings.NewReader(xml))
		if len(errs) > 0 {
			t.Errorf("ufwb.ParseGrammar(%q) = %q want nil error", test.xml, errs)
			continue
		}

		file := ufwb.NewFileFromBytes(binary)
		decoder := ufwb.NewDecoder(g, file)
		got, err := decoder.Decode()
		if err != nil {
			t.Errorf("decoder.Decode() = %q want nil error", err)
			continue
		}

		if err := got.Validiate(); err != nil {
			t.Errorf("value.Validiate() = %q want nil error", err)
		}

		s, err := got.Format(file)
		if err != nil {
			t.Errorf("value.String(...) = %q want nil error", err)
		}
		if s != test.want {
			t.Errorf("value.String(...) = %q want %q", s, test.want)
		}
		//if diff := pretty.Compare(got, test.want); diff != "" {
		//	t.Errorf("decoder.Decode() = -got +want:\n%s", diff)
		//}
	}
}
