package ufwb_test

import (
	"bramp.net/dsector/ufwb"
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/kylelemons/godebug/pretty"
	"io"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
)

const grammarsPath = "../grammars"

func init() {
	// TODO This is bad, move into per test configs
	pretty.CompareConfig.SkipZeroFields = true
	pretty.CompareConfig.IncludeUnexported = false
}

func readGrammar(name string) ([]byte, error) {

	name = path.Join(grammarsPath, name)
	r, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %s", name, err)
	}
	defer r.Close()

	in, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %s", name, err)
	}

	return in, nil
}

type byName []xml.Attr

func (a byName) Len() int {
	return len(a)
}
func (a byName) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a byName) Less(i, j int) bool {
	return a[i].Name.Local < a[j].Name.Local
}

func tokenToString(token xml.Token) string {

	// Type and normalise
	if t, ok := token.(xml.StartElement); ok {
		sort.Sort(byName(t.Attr))
	}

	// There must be a better way than creating an encoder :/
	var out bytes.Buffer
	encoder := xml.NewEncoder(bufio.NewWriter(&out))
	encoder.EncodeToken(token)
	encoder.Flush()

	return string(out.Bytes())
}

func contains(haystack []string, needle string) bool {
	for _, straw := range haystack {
		if needle == straw {
			return true
		}
	}
	return false
}

// nextToken returns the next token skipping whitespace
func nextToken(decoder *xml.Decoder) (xml.Token, error) {
	const skipSpace = true
	const skipComments = true
	skipTags := []string{"fixedvalues"}

	for {
		t, err := decoder.Token()
		if err != nil {
			return t, err
		}

		switch t := t.(type) {
		case xml.CharData:
			if skipSpace && strings.TrimSpace(string(t)) == "" {
				// skip
				continue
			}
		case xml.Comment:
			if skipComments {
				continue
			}
		case xml.StartElement:
			if contains(skipTags, t.Name.Local) {
				continue
			}
		case xml.EndElement:
			if contains(skipTags, t.Name.Local) {
				continue
			}
		}

		return t, err
	}
}

// compareXML compares to XML files and returns the first element to not match
// TODO Move this into its own library
func compareXML(got, want io.Reader) error {
	gotDecoder := xml.NewDecoder(got)
	wantDecoder := xml.NewDecoder(want)

	for {
		gotToken, gotErr := nextToken(gotDecoder)
		wantToken, wantErr := nextToken(wantDecoder)

		if gotErr != nil || wantErr != nil {
			if gotErr == wantErr {
				return nil
			}
			return fmt.Errorf("error got %q want %q", gotErr, wantErr)
		}

		gotS := tokenToString(gotToken)
		wantS := tokenToString(wantToken)
		if gotS != wantS {
			return fmt.Errorf("\n got: %q\nwant: %q", gotS, wantS)
		}
	}
}

/*
            <scriptelement name="JumpToEnd" id="26">
                <script name="unnamed" type="Generic">
                    <source language="Lua">byteView = currentMapper:getCurrentByteView()

fileLength = byteView:getLength()

currentGrammar = currentMapper:getCurrentGrammar()

-- get the structure we want to apply
structure = currentGrammar:getStructureByName(&quot;ZIP end of central directory record&quot;)

bytesProcessed = currentMapper:mapStructureAtPosition(structure, fileLength-22, 22)
</source>
                </script>
            </scriptelement>
*/

func TestParseGrammarFragment(t *testing.T) {

	var tests = []struct {
		xml  string
		want interface{}
	}{
		{
			xml: `<string name="string" id="1">
			        <fixedvalues>
                      <fixedvalue name="up" value="0"/>
					  <fixedvalue name="down" value="1"/>
					  <fixedvalue name="left" value="2"/>
					  <fixedvalue name="right" value="3"/>
                    </fixedvalues>
				</string>`,
			want: ufwb.String{
				XMLName: xml.Name{Local: "string"},
				Name:    "string",
				Id:      1,
				Values: []*ufwb.FixedValue{
					{XMLName: xml.Name{Local: "fixedvalue"}, Name: "up", Value: "0"},
					{XMLName: xml.Name{Local: "fixedvalue"}, Name: "down", Value: "1"},
					{XMLName: xml.Name{Local: "fixedvalue"}, Name: "left", Value: "2"},
					{XMLName: xml.Name{Local: "fixedvalue"}, Name: "right", Value: "3"},
				},
			},
		},{
			xml: `<structure name="structure" id="4" length="prev.size" />`,
			want: ufwb.Structure{
				XMLName: xml.Name{Local: "structure"},
				Name:    "structure",
				Id:      4,
				Length:  "prev.size",
			},
		},{
			xml: `<number name="number" id="3" type="integer" length="4"/>`,
			want: ufwb.Number{
				XMLName: xml.Name{Local: "number"},
				Name:    "number",
				Id:      3,
				Type:    "integer",
				Length:  "4",
			},
		},{
			xml: `<grammar name="multiline">
<description>Line 1
Line 2
Line 3
</description>
			      </grammar>`,
			want: ufwb.Grammar{
				XMLName: xml.Name{Local: "grammar"},
				Name:    "multiline",
				Description: "Line 1\nLine 2\nLine 3\n",
			},
		},{
			xml: `<grammar name="script">
			        <scripts>
			          <script>
			            <description>A description</description>
			            <source language="Python"># Some code</source>
			          </script>
				    </scripts>
			      </grammar>`,

			want: ufwb.Grammar{
				XMLName: xml.Name{Local: "grammar"},
				Name:    "script",
				Scripts: []*ufwb.Script{{
					XMLName: xml.Name{Local: "script"},
					Description: "A description",
					Source: &ufwb.Source{
						XMLName: xml.Name{Local: "source"},
						Language: "Python",
						Text: "# Some code",
					}},
				},
			},
		},
	}

	for i, test := range tests {
		// Reflectively create a instance of the wanted type
		got := reflect.New(reflect.TypeOf(test.want)).Interface()

		decoder := xml.NewDecoder(strings.NewReader(test.xml))
		if err := decoder.Decode(&got); err != nil {
			t.Errorf("Decode(%d) = %s", i, err)
			continue
		}

		if diff := pretty.Compare(got, test.want); diff != "" {
			t.Errorf("Decode(%d) = -got +want:\n%s", i, diff)
			continue
		}

		var out bytes.Buffer
		encoder := xml.NewEncoder(bufio.NewWriter(&out))
		if err := encoder.Encode(&got); err != nil {
			t.Errorf("Encode(%d) = %s", i, err)
			continue
		}
		encoder.Flush()

		if err := compareXML(bytes.NewReader(out.Bytes()), bytes.NewReader([]byte(test.xml))); err != nil {
			t.Errorf("compareXML(%d): %s", i, err)
		}
	}
}

func TestParseGrammar(t *testing.T) {

	var tests = []struct {
		xml  string
		want *ufwb.Ufwb
	}{
		{
			xml: `<ufwb version="1.0.3">
				<grammar name="Test Name" start="1" author="bramp@" fileextension="test" complete="yes">
					<description>Test Description</description>
					<structure name="struct" id="1">
						<string name="string" id="2" type="zero-terminated"/>
						<number name="number" id="3" type="integer" length="4"/>
						<structure name="substruct" id="4" length="prev.size"></structure>
					</structure>
				</grammar>
			</ufwb>`,
			want: &ufwb.Ufwb{
				XMLName: xml.Name{Local: "ufwb"},
				Version: "1.0.3",
				Grammar: &ufwb.Grammar{
					XMLName:     xml.Name{Local: "grammar"},
					Name:        "Test Name",
					Description: "Test Description",
					Author:      "bramp@",
					Ext:         "test",
					Complete:    "yes",
					Start:       "1",
					Structures: []*ufwb.Structure{{
						XMLName: xml.Name{Local: "structure"},
						Name:    "struct",
						Id:      1,
						Elements: []ufwb.Element{
							&ufwb.String{
								XMLName: xml.Name{Local: "string"},
								Name:    "string",
								Id:      2,
								Type:    "zero-terminated",
							},
							&ufwb.Number{
								XMLName: xml.Name{Local: "number"},
								Name:    "number",
								Id:      3,
								Type:    "integer",
								Length:  "4",
							},
							&ufwb.Structure{
								XMLName: xml.Name{Local: "structure"},
								Name:    "substruct",
								Id:      4,
								Length:  "prev.size",
							},
						},
					}},
				},
			},
		}}

	for i, test := range tests {
		got, err := ufwb.ParseGrammar(strings.NewReader(test.xml))
		if err != nil {
			t.Errorf("Parse(%d) = %s", i, err)
			return
		}

		if diff := pretty.Compare(got, test.want); diff != "" {
			t.Errorf("Parse(%d) = -got +want:\n%s", i, diff)
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
		grammar, err := ufwb.ParseGrammar(bytes.NewReader(in))
		if err != nil {
			t.Errorf("Parse(%q) = %s", test, err)
			continue
		}

		// Now write it back and see if it differs
		var out bytes.Buffer
		if err := ufwb.WriteGrammar(bufio.NewWriter(&out), grammar); err != nil {
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
	t.Logf("%v", ufwb.AttrCounter)
}
