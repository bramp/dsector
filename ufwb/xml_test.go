package ufwb

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/kylelemons/godebug/pretty"
	"io"
	"io/ioutil"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func init() {
	// TODO This is bad, move into per test configs
	pretty.CompareConfig.SkipZeroFields = true
	pretty.CompareConfig.IncludeUnexported = false
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if needle == s {
			return true
		}
	}
	return false
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

func TestXmlDecode(t *testing.T) {

	var tests = []struct {
		xml  string
		want XmlElement
	}{
		{
			xml: `<number name="number" id="1" type="integer" length="1">
                    <fixedvalues>
                      <fixedvalue name="name" value="0">
                        <description>some long description</description>
                      </fixedvalue>
                    </fixedvalues>
                  </number>`,
			want: &XmlNumber{
				XMLName:   xml.Name{Local: "number"},
				XmlIdName: XmlIdName{1, "number", ""},
				Length:    "1",
				Type:      "integer",
				Values: []*XmlFixedValue{
					{XMLName: xml.Name{Local: "fixedvalue"},
						Name:        "name",
						Value:       "0",
						Description: "some long description",
					},
				},
			},
		}, {
			xml: `<string name="string" id="1">
			        <fixedvalues>
                      <fixedvalue name="up" value="0"/>
					  <fixedvalue name="down" value="1"/>
					  <fixedvalue name="left" value="2"/>
					  <fixedvalue name="right" value="3"/>
                    </fixedvalues>
				</string>`,
			want: &XmlString{
				XMLName:   xml.Name{Local: "string"},
				XmlIdName: XmlIdName{1, "string", ""},
				Values: []*XmlFixedValue{
					{XMLName: xml.Name{Local: "fixedvalue"}, Name: "up", Value: "0"},
					{XMLName: xml.Name{Local: "fixedvalue"}, Name: "down", Value: "1"},
					{XMLName: xml.Name{Local: "fixedvalue"}, Name: "left", Value: "2"},
					{XMLName: xml.Name{Local: "fixedvalue"}, Name: "right", Value: "3"},
				},
			},
		}, {
			xml: `<structure name="structure" id="4" length="prev.size" />`,
			want: &XmlStructure{
				XMLName:   xml.Name{Local: "structure"},
				XmlIdName: XmlIdName{4, "structure", ""},
				Length:    "prev.size",
			},
		}, {
			xml: `<number name="number" id="3" type="integer" length="4"/>`,
			want: &XmlNumber{
				XMLName:   xml.Name{Local: "number"},
				XmlIdName: XmlIdName{3, "number", ""},
				Type:      "integer",
				Length:    "4",
			},
		}, {
			xml: `<grammar name="multiline">
<description>Line 1
Line 2
Line 3
</description>
			      </grammar>`,
			want: &XmlGrammar{
				XMLName:   xml.Name{Local: "grammar"},
				XmlIdName: XmlIdName{0, "multiline", "Line 1\nLine 2\nLine 3\n"},
			},
		}, {
			xml: `<grammar name="script">
			        <scripts>
			          <script>
			            <description>A description</description>
			            <source language="Python"># Some code</source>
			          </script>
				    </scripts>
			      </grammar>`,

			want: &XmlGrammar{
				XMLName:   xml.Name{Local: "grammar"},
				XmlIdName: XmlIdName{0, "script", ""},
				Scripts: []*XmlScript{{
					XMLName:   xml.Name{Local: "script"},
					XmlIdName: XmlIdName{0, "", "A description"},
					Source: &XmlSource{
						XMLName:  xml.Name{Local: "source"},
						Language: "Python",
						Text:     "# Some code",
					}},
				},
			},
		}, {
			xml: `<scriptelement name="scriptelement">
					<script id="1" name="script" type="Generic">
						<source language="Lua">-- Some code</source>
					</script>
				</scriptelement>`,
			want: &XmlScriptElement{
				XMLName:   xml.Name{Local: "scriptelement"},
				XmlIdName: XmlIdName{0, "scriptelement", ""},
				Script: &XmlScript{
					XMLName:   xml.Name{Local: "script"},
					XmlIdName: XmlIdName{1, "script", ""},
					Type:      "Generic",
					Source: &XmlSource{
						XMLName:  xml.Name{Local: "source"},
						Language: "Lua",
						Text:     "-- Some code",
					},
				},
			},
		},
	}

	for i, test := range tests {
		// Reflectively create a instance of the wanted type
		got := reflect.New(reflect.TypeOf(test.want)).Interface()

		decoder := xml.NewDecoder(strings.NewReader(test.xml))
		if err := decoder.Decode(got); err != nil {
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

// Decode all grammars xml, but does not check if it is valid.
func TestXmlDecodeAll(t *testing.T) {
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

		// Decode
		got := &XmlUfwb{}
		decoder := xml.NewDecoder(bytes.NewReader(in))
		if err := decoder.Decode(got); err != nil {
			t.Errorf("Decode(%q) = %q", test, err)
			continue
		}

		// Now write it back and see if it differs
		var out bytes.Buffer
		w := bufio.NewWriter(&out)
		w.Write([]byte(xml.Header))
		encoder := xml.NewEncoder(w)
		if err := encoder.Encode(&got); err != nil {
			t.Errorf("Encode(%q) = %s", test, err)
			continue
		}

		if err := compareXML(bytes.NewReader(out.Bytes()), bytes.NewReader(in)); err != nil {
			t.Errorf("compareXML(%q): %s", test, err)
			continue
		}

		good++
	}

	if found == 0 {
		t.Fatalf("Failed to find any grammars")
	}

	t.Logf("Progress: %d/%d good", good, found)
	t.Logf("%v", AttrCounter)
}
