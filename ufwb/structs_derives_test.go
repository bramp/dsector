package ufwb

import (
	"github.com/kylelemons/godebug/pretty"
	"strings"
	"testing"
)

func TestExtends(t *testing.T) {
	// A more complex extends example (modified from a real packet capture).
	xml := testHeader +
		`<structure name="packets" id="99" order="variable">
			<structref name="IPv4 TCP Packet" repeatmin="0" structure="tcp4"/>
			<structref name="IPv4 Packet" repeatmin="0" structure="ip4"/>
			<structref name="Ethernet Frame" repeatmin="0" structure="eth"/>
		</structure>

		<structure name="eth" id="2">
			<binary name="eth_dst" id="5" length="6"/>
			<binary name="eth_src" id="6" length="6"/>
			<number name="eth.type" id="7" mustmatch="yes" type="integer" length="2">
				<fixedvalues>
					<fixedvalue name="IPv4 Packet" value="0x800"/>
					<fixedvalue name="ARP Frame" value="0x806"/>
					<fixedvalue name="IPv6 Packet" value="0x86DD"/>
				</fixedvalues>
			</number>
		</structure>

		<structure name="ip4" id="3" extends="id:2">
			<binary name="eth_dst" id="8"/>
			<binary name="eth_src" id="9"/>
			<number name="eth.type" id="10" type="integer">
				<fixedvalues>
					<!-- This should replace the 3 possible options -->
					<fixedvalue name="IPv4 Packet" value="0x800"/>
				</fixedvalues>
			</number>
			<!-- TODO 4 bit numbers are not yet supported :( -->
			<number name="ip.version" id="11" type="integer" length="4" lengthunit="bit"/>
			<number name="ip.hdr_len" id="12" type="integer" length="4" lengthunit="bit"/>
			<number name="ip.proto" mustmatch="yes" id="13" type="integer" length="1"/>
		</structure>

		<structure name="tcp4" id="4" extends="id:3">
			<number name="ip.version" id="14" type="integer"/>
			<number name="ip.proto" id="15" type="integer">
				<fixedvalues>
					<fixedvalue name="TCP" value="6"/>
				</fixedvalues>
			</number>
			<number name="tcp.srcport" id="16" type="integer" length="2"/>
			<number name="tcp.dstport" id="17" type="integer" length="2"/>
		</structure>` + testFooter

	wantNames := []string{
		"eth_dst", "eth_src", "eth.type",
		"ip.version", "ip.hdr_len", "ip.proto",
		"tcp.srcport", "tcp.dstport"}
	wantEthProto := []string{"IPv4 Packet"}
	wantIpProto := []string{"TCP"}

	grammar, errs := ParseXmlGrammar(strings.NewReader(xml))
	if len(errs) > 0 {
		t.Errorf("ParseXmlGrammar(...) = %q want nil error", errs)
		return
	}

	e, found := grammar.Get("tcp4")
	if !found {
		t.Errorf("grammar.Get(\"tcp4\") not found, want found")
		return
	}

	s, ok := e.(*Structure)
	if !ok {
		t.Errorf("grammar.Get(\"tcp4\") = %T, want *Structure", e)
		return
	}

	// Get the names of all the fields within the structure
	var names []string
	elements := s.Elements()
	for _, element := range elements {
		names = append(names, element.Name())
	}

	if diff := pretty.Compare(names, wantNames); diff != "" {
		t.Errorf("grammar.Get(\"tcp4\") = -got +want:\n%s", diff)
	}

	// Get the "eth.type" field, which should be fixed to "IPv4 Packet"
	n, ok := elements[2].(*Number)
	if !ok {
		t.Errorf("grammar.Elements()['eth.type'] = %T, want *Number", e)
		return
	}

	// Get the name of each value
	names = nil
	for _, value := range n.Values() {
		names = append(names, value.name)
	}

	if diff := pretty.Compare(names, wantEthProto); diff != "" {
		t.Errorf("grammar.Elements()['eth.type'] = -got +want:\n%s", diff)
	}

	// Get the "ip.proto" field, which should be fixed to "TCP"
	n, ok = elements[5].(*Number)
	if !ok {
		t.Errorf("grammar.Elements()['ip.proto'] = %T, want *Number", e)
		return
	}

	// Get the name of each value
	names = nil
	for _, value := range n.Values() {
		names = append(names, value.name)
	}

	if diff := pretty.Compare(names, wantIpProto); diff != "" {
		t.Errorf("grammar.Elements()['ip.proto'] = -got +want:\n%s", diff)
	}
}
