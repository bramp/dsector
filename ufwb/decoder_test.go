package ufwb

import (
	"bramp.net/dsector/input"
	"strings"
	"testing"
)

func TestDecoderPrev(t *testing.T) {
	grammar := `<ufwb version="1.0.3">
					<grammar start="1">
						<structure name="struct" id="1" repeatmax="unlimited">
							<string name="a" id="2" type="zero-terminated"/>
							<string name="b" id="3" type="zero-terminated"/>
							<string name="c" id="4" type="zero-terminated"/>
						</structure>
					</grammar>
				</ufwb>`
	data := []byte{'A', 0, 'B', 0, 'C', 0, 'D', 0, 'E', 0, 'F', 0}
	want := map[string]string{
		"a": "D",
		"b": "E",
		"c": "F",
	}
	file := input.FromBytes(data)

	ufwb, errs := ParseXmlGrammar(strings.NewReader(grammar))
	if len(errs) > 0 {
		t.Fatalf("ParseXmlGrammar(...) = %s, want nil", errs)
	}

	decoder := NewDecoder(ufwb, file)
	_, err := decoder.Decode()
	if err != nil {
		t.Fatalf("decoder.Decode() = %s, want nil", err)
	}

	for name, letter := range want {
		v, err := decoder.prevByName(name)
		if err != nil {
			t.Errorf("decoder.prev(%q) = %q, want nil", name, err)
		}

		got, err := v.Element.(*String).Format(file, v)
		if err != nil {
			t.Errorf("v.Format(%q) = %q, want nil", name, err)
		}
		if got != letter {
			t.Errorf("v.Format(%q) = %q, want %q", name, got, letter)
		}
	}
}
