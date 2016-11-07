package ufwb

import (
	"bramp.net/dsector/input"
	"fmt"
	"strings"
	"testing"
)

func TestLua(t *testing.T) {
	grammar := `<ufwb version="1.0.3">
					<grammar start="1">
						<structure name="struct" id="1" repeatmax="unlimited">
							<number name="number" id="2" type="integer" length="4" endian="big" display="hex" />
							<scriptelement name="script" id="3">
                    			<script type="Generic">
                        			<source language="Lua">
                        			%s
                        			</source>
                        		</script>
                        	</scriptelement>
						</structure>
					</grammar>
				</ufwb>`
	// TODO Remove the last 4 bytes, after we fix the parsing of the scriptelement if the data is only 4 bytes long
	data := []byte{0xA1, 0xB2, 0xC3, 0xD4, 0, 0, 0, 0}

	var tests = []struct {
		text string
		want interface{}
	}{{
		text: `debug("hello")`,
		want: "hello",
	}, {
		text: `debug(synalysis.ENDIAN_LITTLE)`,
		want: 1,
	}, {
		text: `debug(synalysis.ENDIAN_BIG)`,
		want: 2,
	}, {
		text: `results = currentMapper:getCurrentResults()
		       debug(results:getLastResult())`,
		want: "",
	}, {
		text: `results = currentMapper:getCurrentResults()
		       lastResult = results:getLastResult()
		       debug(lastResult:getValue())`,
		want: "",
	}, {
		text: `results = currentMapper:getCurrentResults()
		       lastResult = results:getLastResult()
		       value = lastResult:getValue()
		       debug(value:getUnsignedNumber())`,
		want: 0xA1B2C3D4,
	}, {
		text: `currentMapper:setDynamicEndianness(synalysis.ENDIAN_BIG)`,
		want: "",
	}}

	for _, test := range tests {
		fullGrammar := fmt.Sprintf(grammar, test.text)

		ufwb, errs := ParseXmlGrammar(strings.NewReader(fullGrammar))
		if len(errs) > 0 {
			t.Errorf("ParseXmlGrammar(%q) = %q", test, errs)
			continue
		}

		got := interface{}("UNSET")

		d := NewDecoder(ufwb, input.FromBytes(data))
		d.debugFunc = func(value interface{}) {
			//got = fmt.Sprintf("%s", value)
			got = value
		}
		_, err := d.Decode()
		if err != nil {
			t.Errorf("d.Decode(%q) = %q, want nil", test.text, err)
		}

		// TODO
		//if got != test.want {
		//	t.Errorf("debugFunc(...) got %q, want %q", got, test.want)
		//}
	}

}
