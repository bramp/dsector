package ufwb

import (
	"bramp.net/dsector/input"
	"fmt"
	"github.com/yuin/gopher-lua"
	"strings"
	"testing"
)

func TestLua(t *testing.T) {
	grammar := `<ufwb version="1.0.3">
					<grammar start="1">
						<structure name="struct" id="1" repeatmax="unlimited">
							<number name="number" id="2" type="integer" length="4" endian="big" display="hex" signed="no" />
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

	data := []byte{0xA1, 0xB2, 0xC3, 0xD4}

	var tests = []struct {
		text string
		want interface{}
	}{{
		text: `debug("hello")`,
		want: "hello",
	}, {
		text: `debug(synalysis.ENDIAN_BIG)`,
		want: lua.LNumber(2),
	}, {
		text: `debug(synalysis.ENDIAN_LITTLE)`,
		want: lua.LNumber(3),
	}, {
		text: `results = currentMapper:getCurrentResults()
		       debug(results:getLastResult())`,
		want: nil,
	}, {
		text: `results = currentMapper:getCurrentResults()
		       lastResult = results:getLastResult()
		       debug(lastResult:getValue())`,
		want: nil,
	}, {
		text: `results = currentMapper:getCurrentResults()
		       lastResult = results:getLastResult()
		       value = lastResult:getValue()
		       debug(value:getUnsignedNumber())`,
		want: lua.LNumber(0xA1B2C3D4),
	}, {
		text: `debug(currentMapper:getDynamicEndianness())`,
		want: lua.LNumber(2), // Default is ENDIAN_BIG
	}, {
		text: `currentMapper:setDynamicEndianness(synalysis.ENDIAN_LITTLE)
		       debug(currentMapper:getDynamicEndianness())`,
		want: lua.LNumber(3),
	}}

	for _, test := range tests {
		fullGrammar := fmt.Sprintf(grammar, test.text)

		ufwb, errs := ParseXmlGrammar(strings.NewReader(fullGrammar))
		if len(errs) > 0 {
			t.Errorf("ParseXmlGrammar(%q) = %q", test, errs)
			continue
		}

		got := interface{}("<nothing>")

		d := NewDecoder(ufwb, input.FromBytes(data))
		d.debugFunc = func(value interface{}) {
			got = value
		}
		_, err := d.Decode()
		if err != nil {
			t.Errorf("d.Decode(%q) = %q, want nil", test.text, err)
		}

		if test.want != nil && got != test.want {
			t.Errorf("debugFunc(...) got (%T)%q, want %q", got, got, test.want)
		}
	}

}
