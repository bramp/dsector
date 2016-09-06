package ufwb_test

import (
	"bramp.net/dsector/ufwb"
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
				Version: "1.0.3",
				Grammar: &ufwb.Grammar{
					IdName:   ufwb.IdName{0, "Test Name", "Test Description"},
					Author:   "bramp@",
					Ext:      "test",
					Complete: true,
					//Start:       "1",
					Elements: []ufwb.Element{
						&ufwb.Structure{
							IdName: ufwb.IdName{1, "struct", ""},
							Elements: []ufwb.Element{
								&ufwb.String{
									IdName: ufwb.IdName{2, "string", ""},
									Type:   "zero-terminated",
								},
								&ufwb.Number{
									IdName: ufwb.IdName{3, "number", ""},
									Type:   "integer",
									Length: "4",
								},
								&ufwb.Structure{
									IdName: ufwb.IdName{4, "substruct", ""},
									Length: "prev.size",
								},
							},
						},
					},
				},
			},
		}}

	for i, test := range tests {
		got, err := ufwb.ParseXmlGrammar(strings.NewReader(test.xml))
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
		grammar, errs := ufwb.ParseXmlGrammar(bytes.NewReader(in))
		if len(errs) > 0 {
			t.Errorf("Parse(%q) = %q", test, errs)
			continue
		}

		// Now write it back and see if it differs
		var out bytes.Buffer
		if err := ufwb.WriteXmlGrammar(bufio.NewWriter(&out), grammar); err != nil {
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
