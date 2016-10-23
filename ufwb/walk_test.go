package ufwb_test

import (
	"bramp.net/dsector/ufwb"
	"github.com/kylelemons/godebug/pretty"
	"strings"
	"testing"
	"bramp.net/dsector/toerr"
)

func TestWalk(t *testing.T) {

	xml := `<ufwb>
				<grammar name="Test Name" start="1">
					<description>Test Description</description>
					<structure name="struct" id="1">
						<string name="string" id="2" type="zero-terminated"/>
						<number name="number" id="3" type="integer" length="4"/>
						<structure name="substruct" id="4" length="prev.size"></structure>
					</structure>
				</grammar>
			</ufwb>`

	g, err := ufwb.ParseXmlGrammar(strings.NewReader(xml))
	if err != nil {
		t.Fatalf("ParseXmlGrammar(...) = %q want nil", err)
	}

	type walkResult struct {
		Id       int
		Name     string
		ParentId int
	}

	var got []walkResult

	errs := ufwb.Walk(g, func(root *ufwb.Ufwb, element ufwb.Element, parent *ufwb.Structure, errs *toerr.Errors) {
		parentId := 0
		if parent != nil {
			parentId = parent.Id()
		}
		got = append(got, walkResult{
			Id:       element.Id(),
			Name:     element.Name(),
			ParentId: parentId,
		})
	})

	if len(errs) != 0 {
		t.Errorf("Walk(...) = %q want nil", errs)
	}

	want := []walkResult{
		{0, "Test Name", 0},
		{1, "struct", 0},
		{2, "string", 1},
		{3, "number", 1},
		{4, "substruct", 1},
	}

	if diff := pretty.Compare(got, want); diff != "" {
		t.Errorf("Walk(...) = -got +want:\n%s", diff)
	}

}
