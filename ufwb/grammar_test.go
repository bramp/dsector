package ufwb

import (
	"bramp.net/dsector/toerr"
	"bufio"
	"bytes"
	"fmt"
	"github.com/kylelemons/godebug/pretty"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
)

const grammarsPath = "../grammars"

var config *pretty.Config

func init() {
	config = &pretty.Config{
		IncludeUnexported:   true,
		PrintStringers:      false,
		PrintTextMarshalers: false,
		SkipZeroFields:      true,
	}
}

// readGrammar from filename
func readGrammar(filename string) ([]byte, error) {
	filename = path.Join(grammarsPath, filename)
	r, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %s", filename, err)
	}
	defer r.Close()

	in, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %s", filename, err)
	}

	return in, nil
}

// normalise strips XML, and all pointers from the Ufwb. This is to avoid loops, which
// confuse the pretty.Compare(...).
func normalise(root *Ufwb, element Element, parent *Structure, errs *toerr.Errors) {
	switch e := element.(type) {
	case *Grammar:
		e.Xml = nil
		e.Start = nil
	case *Structure:
		e.Xml = nil
		e.extends = nil
		e.parent = nil
	case *GrammarRef:
		e.Xml = nil
		e.extends = nil
	case *Custom:
		e.Xml = nil
		e.extends = nil
	case *StructRef:
		e.Xml = nil
		e.extends = nil
	case *String:
		e.Xml = nil
		e.extends = nil
		e.parent = nil
		for _, v := range e.values {
			v.Xml = nil
			v.extends = nil
		}
	case *Binary:
		e.Xml = nil
		e.extends = nil
		e.parent = nil
		for _, v := range e.values {
			v.Xml = nil
			v.extends = nil
		}
	case *Number:
		e.Xml = nil
		e.extends = nil
		e.parent = nil
		for _, v := range e.values {
			v.Xml = nil
			v.extends = nil
		}
	case *Offset:
		e.Xml = nil
		e.extends = nil
	default:
		errs.Append(fmt.Errorf("unknown element type %T", element))
	}
}

func TestParseGrammarFragment(t *testing.T) {

	var tests = []struct {
		xml  string
		want Element
	}{
		{
			xml: `<number name="number name" id="1" type="integer" length="1">
                    <fixedvalues>
                        <fixedvalue name="first value" value="0">
                            <description>
                            Some description
                        </description>
                        </fixedvalue>
                    </fixedvalues>
                </number>`,
			want: &Number{
				Base:   Base{"Number", 1, "number name", ""},
				Type:   "integer",
				length: "1",
				values: []*FixedValue{
					{name: "first value", value: 0},
				},
			},
		},
	}

	for i, test := range tests {
		xml := commonHeader + test.xml + commonFooter
		got, errs := ParseXmlGrammar(strings.NewReader(xml))
		if len(errs) > 0 {
			t.Errorf("Parse(test:%d) = %s", i, errs)
			continue
		}

		// Remove all the XML fields, as we don't want to compare them
		start := got.Grammar.Start
		errs = Walk(got, normalise)
		if len(errs) > 0 {
			t.Errorf("Walk(test:%d) errors: %s", i, errs)
			continue
		}

		if diff := config.Compare(start, test.want); diff != "" {
			t.Errorf("Parse(test:%d) = -got +want:\n%s", i, diff)
		}
	}
}

func TestParseGrammar(t *testing.T) {

	var tests = []struct {
		xml  string
		want *Ufwb
	}{
		{
			xml: `<ufwb version="1.0.3">
				<grammar name="Test Name" start="1" author="bramp@" fileextension="test" complete="yes">
					<description>Test Description</description>
					<structure name="struct" id="1">
						<string name="string" id="2" type="zero-terminated"/>
						<number name="number" id="3" type="integer" length="8"/>
						<structure name="substruct" id="4" length="prev.number">
							<binary name="binary" id="5" length="4">
								<fixedvalues>
									<fixedvalue name="one" value="0x01234567"/>
									<fixedvalue name="two" value="0x89abcdef"/>
								</fixedvalues>
							</binary>
							<number name="number_values" id="6" type="integer" length="4">
								<fixedvalues>
									<fixedvalue name="three" value="0xfedcba98"/>
									<fixedvalue name="four"  value="0x76543210"/>
								</fixedvalues>
							</number>
						</structure>
					</structure>
				</grammar>
			</ufwb>`,
			want: &Ufwb{
				Version: "1.0.3",
				Grammar: &Grammar{
					Base:     Base{"Grammar", 0, "Test Name", "Test Description"},
					Author:   "bramp@",
					Ext:      "test",
					Complete: boolOf(true),
					//Start:       "1",
					Elements: []Element{
						&Structure{
							Base: Base{"Structure", 1, "struct", ""},
							elements: []Element{
								&String{
									Base: Base{"String", 2, "string", ""},
									typ:  "zero-terminated",
								},
								&Number{
									Base:   Base{"Number", 3, "number", ""},
									Type:   "integer",
									length: "8",
								},
								&Structure{
									Base:   Base{"Structure", 4, "substruct", ""},
									length: "prev.number",
									elements: []Element{
										&Binary{
											Base:   Base{"Binary", 5, "binary", ""},
											length: "4",
											values: []*FixedBinaryValue{
												{name: "one", value: []byte{0x01, 0x23, 0x45, 0x67}},
												{name: "two", value: []byte{0x89, 0xab, 0xcd, 0xef}},
											},
										},
										&Number{
											Base:   Base{"Number", 6, "number_values", ""},
											Type:   "integer",
											length: "4",
											values: []*FixedValue{
												{name: "three", value: 0xfedcba98},
												{name: "four", value: 0x76543210},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}}

	for i, test := range tests {
		got, errs := ParseXmlGrammar(strings.NewReader(test.xml))
		if len(errs) > 0 {
			t.Errorf("Parse(test:%d) = %s", i, errs)
			continue
		}

		// Remove all the XML fields, as we don't want to compare them
		got.Xml = nil
		got.Elements = nil
		errs = Walk(got, normalise)
		if len(errs) > 0 {
			t.Errorf("Walk(test:%d) errors: %s", i, errs)
			continue
		}

		if diff := config.Compare(got, test.want); diff != "" {
			t.Errorf("Parse(test:%d) = -got +want:\n%s", i, diff)
		}
	}
}

// Open all grammars and see if we can parse
// TODO Make this test pass
func TestParserAll(t *testing.T) {
	files, err := ioutil.ReadDir(grammarsPath)
	if err != nil {
		t.Fatalf("Failed to read grammar directory: %s", err)
	}

	var found, good int
	for _, file := range files {

		test := file.Name()
		if path.Ext(test) != ".grammar" || file.IsDir() {
			continue
		}
		found++

		in, err := readGrammar(test)
		if err != nil {
			t.Errorf("readGrammar(%s) = %s", test, err)
			continue
		}

		// Parse
		grammar, errs := ParseXmlGrammar(bytes.NewReader(in))
		if len(errs) > 0 {
			t.Errorf("Parse(%q) = %q", test, errs)
			continue
		}

		// Now write it back and see if it differs
		var out bytes.Buffer
		if err := WriteXmlGrammar(bufio.NewWriter(&out), grammar); err != nil {
			t.Errorf("Write(%q) = %s", test, err)
			continue
		}

		ioutil.WriteFile(path.Join(grammarsPath, test)+".test", out.Bytes(), 0777)

		if err := compareXML(bytes.NewReader(out.Bytes()), bytes.NewReader(in)); err != nil {
			t.Errorf("compareXML(%q): %s", test, err)
			continue
		}

		good++
	}

	if found == 0 {
		t.Fatalf("Failed to find any grammars")
	}

	// TODO Got to Progress: 67/79 good
	t.Logf("Progress: %d/%d good", good, found)
	t.Logf("%v", AttrCounter)
}
