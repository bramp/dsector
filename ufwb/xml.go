package ufwb

import (
	"encoding/xml"
	"bytes"
	"fmt"
)

type Endian int

const (
	UnknownEndian Endian = iota
	LittleEndian
	BigEndian
	DynamicEndian
)

type Display int

const (
	UnknownDisplay Display = iota
	BinaryDisplay
	DecDisplay
	HexDisplay
)

func (d Display) Base() int {
	switch d {
	case HexDisplay: return 16
	case DecDisplay: return 10
	case BinaryDisplay: return 2
	case UnknownDisplay: return 0
	}
	return 0
}

type LengthUnit int

const (
	UnknownLengthUnit LengthUnit = iota
	ByteLengthUnit
	BitLengthUnit
)


type Ufwb struct {
	XMLName xml.Name `xml:"ufwb"`

	Version string  `xml:"version,attr,omitempty"`
	Grammar *Grammar `xml:"grammar"`

	// Extra info not encoded in the XML
	elements map[string]Element
}

type Grammar struct {
	XMLName xml.Name `xml:"grammar"`

	Name        string `xml:"name,attr"`
	Description string `xml:"description,omitempty"`
	Author      string `xml:"author,attr,omitempty"`
	Ext         string `xml:"fileextension,attr,omitempty"`
	Email       string `xml:"email,attr,omitempty"`
	Complete    string `xml:"complete,attr,omitempty"`
	Uti         string `xml:"uti,attr,omitempty"`

	Start       string       `xml:"start,attr,omitempty" ufwb:"id"`
	Scripts     Scripts      `xml:"scripts"`
	Structures  []*Structure `xml:"structure,omitempty"`

	start Element // TODO Is this always a Structure?
}

type GrammarRef struct {
	XMLName xml.Name `xml:"grammarref"`

	Id          int    `xml:"id,attr,omitempty"`
	Name        string `xml:"name,attr"`
	Description string `xml:"description,omitempty"`
	Uti         string `xml:"uti,attr,omitempty"`

	Filename     string `xml:"filename,attr,omitempty"`
	Disabled     string `xml:"disabled,attr,omitempty" ufwb:"bool"`
}

type length struct {
	Length      string
	LengthUnit  LengthUnit
}

type numberEncoding struct {
	Endian    Endian
	Signed    bool
}

type display struct {
	Display   Display

	FillColour   uint32
	StrokeColour uint32
}

type Structure struct {
	XMLName xml.Name `xml:"structure"`

	Id          int    `xml:"id,attr,omitempty"`
	Name        string `xml:"name,attr,omitempty"`
	Description string `xml:"description,omitempty"`

	Length       string `xml:"length,attr,omitempty" ufwb:"ref"`
	LengthOffset string `xml:"lengthoffset,attr,omitempty"`

	Endian    string `xml:"endian,attr,omitempty" ufwb:"endian"`
	Signed    string `xml:"signed,attr,omitempty" ufwb:"bool"`
	Extends   string `xml:"extends,attr,omitempty" ufwb:"id"`
	Order     string `xml:"order,attr,omitempty"` // ??
	Encoding  string `xml:"encoding,attr,omitempty" ufwb:"encoding"`
	Alignment string `xml:"alignment,attr,omitempty"` // ??

	Floating   string `xml:"floating,attr,omitempty"` // ??
	ConsistsOf string `xml:"consists-of,attr,omitempty" ufwb:"id"`

	Repeat    string `xml:"repeat,attr,omitempty" ufwb:"id"`
	RepeatMin string `xml:"repeatmin,attr,omitempty" ufwb:"ref"`
	RepeatMax string `xml:"repeatmax,attr,omitempty" ufwb:"ref"`

	ValueExpression string `xml:"valueexpression,attr,omitempty"`
	Debug           string `xml:"debug,attr,omitempty" ufwb:"bool"`
	Disabled        string `xml:"disabled,attr,omitempty" ufwb:"bool"`

	FillColour   string `xml:"fillcolor,attr,omitempty" ufwb:"colour"`
	StrokeColour string `xml:"strokecolor,attr,omitempty" ufwb:"colour"`

	Elements []Element `xml:",any"`

	display   Display
	endian    Endian
	signed    bool
	extends   *Structure // TODO Is this always a struct?
}

type Custom struct {
	XMLName xml.Name `xml:"custom"`

	Id          int    `xml:"id,attr,omitempty"`
	Name        string `xml:"name,attr,omitempty"`
	Description string `xml:"description,omitempty"`

	Length string `xml:"length,attr,omitempty" ufwb:"ref"`
	Script string `xml:"script,attr,omitempty" ufwb:"id"`

	FillColour   string `xml:"fillcolor,attr,omitempty" ufwb:"colour"`
	StrokeColour string `xml:"strokecolor,attr,omitempty" ufwb:"colour"`
}

type StructRef struct {
	XMLName xml.Name `xml:"structref"`

	Id          int    `xml:"id,attr,omitempty"`
	Name        string `xml:"name,attr,omitempty"`
	Description string `xml:"description,omitempty"`

	Structure string `xml:"structure,attr,omitempty" ufwb:"id"`
	RepeatMin string `xml:"repeatmin,attr,omitempty" ufwb:"ref"`
	RepeatMax string `xml:"repeatmax,attr,omitempty" ufwb:"ref"`

	FillColour   string `xml:"fillcolor,attr,omitempty" ufwb:"colour"`
	StrokeColour string `xml:"strokecolor,attr,omitempty" ufwb:"colour"`

	structure *Structure
}

type String struct {
	XMLName xml.Name `xml:"string"`

	Id          int    `xml:"id,attr,omitempty"`
	Name        string `xml:"name,attr,omitempty"`
	Description string `xml:"description,omitempty"`

	Type string `xml:"type,attr,omitempty" ufwb:"string-type"` // "zero-terminated", "fixed-length"

	Length    string `xml:"length,attr,omitempty" ufwb:"ref"`
	Encoding  string `xml:"encoding,attr,omitempty" ufwb:"encoding"` // Should be valid encoding
	MustMatch string `xml:"mustmatch,attr,omitempty" ufwb:"bool"`    // "yes", "no"

	RepeatMin string `xml:"repeatmin,attr,omitempty" ufwb:"ref"`
	RepeatMax string `xml:"repeatmax,attr,omitempty" ufwb:"ref"`

	FillColour   string `xml:"fillcolor,attr,omitempty" ufwb:"colour"`
	StrokeColour string `xml:"strokecolor,attr,omitempty" ufwb:"colour"`

	Values []*FixedValue `xml:"fixedvalue,omitempty"`
}

type Binary struct {
	XMLName xml.Name `xml:"binary"`

	Id          int    `xml:"id,attr,omitempty"`
	Name        string `xml:"name,attr,omitempty"`
	Description string `xml:"description,omitempty"`

	Length     string `xml:"length,attr,omitempty" ufwb:"ref"`
	LengthUnit string `xml:"lengthunit,attr,omitempty" ufwb:"lengthunit"` // "bit"

	RepeatMin string `xml:"repeatmin,attr,omitempty" ufwb:"ref"`
	RepeatMax string `xml:"repeatmax,attr,omitempty" ufwb:"ref"`

	MustMatch string `xml:"mustmatch,attr,omitempty"  ufwb:"bool"`
	Unused    string `xml:"unused,attr,omitempty" ufwb:"bool"`
	Disabled  string `xml:"disabled,attr,omitempty" ufwb:"bool"`

	FillColour   string `xml:"fillcolor,attr,omitempty" ufwb:"colour"`
	StrokeColour string `xml:"strokecolor,attr,omitempty" ufwb:"colour"`

	Values []*FixedValue `xml:"fixedvalue,omitempty"`
}

type Number struct {
	XMLName xml.Name `xml:"number"`

	Id          int    `xml:"id,attr,omitempty"`
	Name        string `xml:"name,attr,omitempty"`
	Description string `xml:"description,omitempty"`

	RepeatMin string `xml:"repeatmin,attr,omitempty" ufwb:"ref"`
	RepeatMax string `xml:"repeatmax,attr,omitempty" ufwb:"ref"`

	Type       string `xml:"type,attr,omitempty" ufwb:"number-type"`
	Length     string `xml:"length,attr,omitempty" ufwb:"ref"`
	LengthUnit string `xml:"lengthunit,attr,omitempty" ufwb:"lengthunit"` // "", "bit" (default "byte")

	Endian          string `xml:"endian,attr,omitempty" ufwb:"endian"` // "", "big", "little", "dynamic"
	Signed          string `xml:"signed,attr,omitempty" ufwb:"bool"`   // "", "yes", "no"
	MustMatch       string `xml:"mustmatch,attr,omitempty" ufwb:"bool"`
	ValueExpression string `xml:"valueexpression,attr,omitempty"`

	MinVal string `xml:"minval,attr,omitempty" ufwb:"ref"`
	MaxVal string `xml:"maxval,attr,omitempty" ufwb:"ref"`

	Display      string `xml:"display,attr,omitempty" ufwb:"display"`
	FillColour   string `xml:"fillcolor,attr,omitempty" ufwb:"colour"`
	StrokeColour string `xml:"strokecolor,attr,omitempty" ufwb:"colour"`

	Disabled string `xml:"disabled,attr,omitempty" ufwb:"bool"`

	Values []*FixedValue `xml:"fixedvalue,omitempty"`
	Masks  []*Mask       `xml:"mask,omitempty"`

	endian    Endian
	signed    bool
	display   Display
}

type Offset struct {
	XMLName xml.Name `xml:"offset"`

	Id          int    `xml:"id,attr,omitempty"`
	Name        string `xml:"name,attr,omitempty"`
	Description string `xml:"description,omitempty"`

	RepeatMin string `xml:"repeatmin,attr,omitempty" ufwb:"ref"`
	RepeatMax string `xml:"repeatmax,attr,omitempty" ufwb:"ref"`

	Length              string `xml:"length,attr,omitempty" ufwb:"ref"`
	Endian              string `xml:"endian,attr,omitempty" ufwb:"endian"`
	RelativeTo          string `xml:"relative-to,attr,omitempty" ufwb:"id"`
	FollowNullReference string `xml:"follownullreference,attr,omitempty"`
	References          string `xml:"references,attr,omitempty" ufwb:"id"`
	ReferencedSize      string `xml:"referenced-size,attr,omitempty" ufwb:"id"`
	Additional          string `xml:"additional,attr,omitempty"` // "stringOffset"

	Display      string `xml:"display,attr,omitempty" ufwb:"display"` // "", "hex", "offset"
	FillColour   string `xml:"fillcolor,attr,omitempty" ufwb:"colour"`
	StrokeColour string `xml:"strokecolor,attr,omitempty" ufwb:"colour"`

	endian    Endian
	display   Display
}

type Mask struct {
	XMLName xml.Name `xml:"mask"`

	Name        string `xml:"name,attr,omitempty"`
	Description string `xml:"description,omitempty"`

	Value  string       `xml:"value,attr,omitempty"`
	Values []*FixedValue `xml:"fixedvalue,omitempty"`
}

type FixedValues struct {
	XMLName xml.Name `xml:"fixedvalues"`

	Values []*FixedValue `xml:"fixedvalue,omitempty"`
}

type FixedValue struct {
	XMLName xml.Name `xml:"fixedvalue"`

	Name        string `xml:"name,attr,omitempty"`
	Description string `xml:"description,omitempty"`

	Value string `xml:"value,attr"`
}

type Scripts []*Script

type ScriptElement struct {
	XMLName xml.Name `xml:"scriptelement"`

	Id   int    `xml:"id,attr,omitempty"`
	Name string `xml:"name,attr,omitempty"`

	Disabled string `xml:"disabled,attr,omitempty" ufwb:"bool"`

	Script *Script `xml:"script"`
}

type Script struct {
	XMLName xml.Name `xml:"script"`

	Id   int    `xml:"id,attr,omitempty"`
	Name string `xml:"name,attr,omitempty"`
	Description string `xml:"description,omitempty"`

	Type string `xml:"type,attr,omitempty"`
	Language string `xml:"language,attr,omitempty" ufwb:"lang"`
	Source   *Source `xml:"source,omitempty"`

	// Sometimes the text is defined here, or in the child Source element
	Text     string `xml:",chardata"` // TODO Should this be cdata?

	source   *Source
}

type Source struct {
	XMLName xml.Name `xml:"source"`

	Language string `xml:"language,attr,omitempty" ufwb:"lang"`
	Text     string `xml:",chardata"` // TODO Should this be cdata?

	language string
}

// Types of the original elements but without the MarshalXML / UnmarshalXML methods on them.
type nakedStructure Structure
type nakedString String
type nakedNumber Number
type nakedBinary Binary

func unmarshalStartElement(v interface{}, start xml.StartElement) error {
	// Because we implement our own UnmarshalXML, the attributes from the StartElement are not
	// copied into the receiver. We have a nasty hack to fix that, by Encoding the StartElement
	// and then Decoding directly into the structure via a type conversion.
	var tag bytes.Buffer
	encoder := xml.NewEncoder(&tag)
	encoder.EncodeToken(start)
	encoder.EncodeToken(start.End())
	encoder.Flush()

	// Now decode the StartElement into the structure.
	in := bytes.NewReader(tag.Bytes())
	return xml.NewDecoder(in).Decode(v)
}

func marshalStartElement(v interface{}) (xml.StartElement, error) {
	// Encode the value, and then read back just the first token
	var tag bytes.Buffer
	e := xml.NewEncoder(&tag)
	err := e.Encode(v)
	if err != nil {
		return xml.StartElement{}, err
	}

	d := xml.NewDecoder(bytes.NewReader(tag.Bytes()))
	start, err := d.Token()

	// We assume the first token is a StartElement
	return start.(xml.StartElement), err
}

func (s *Scripts) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		switch token := token.(type) {
		case xml.StartElement:
			if token.Name.Local == "script" {
				element := &Script{}
				if err = decoder.DecodeElement(element, &token); err != nil {
					return err
				}
				if element.Source != nil {
					// If there is a source element, then the Text shouldn't be set
					element.Text = ""
				}
				*s = append(([]*Script)(*s), element)
			} else {
				return fmt.Errorf("unknown element: `%s` inside `%s` at %d",
					token.Name.Local, start.Name.Local, decoder.InputOffset())
			}
		case xml.EndElement:
			return nil
		}
	}
}

func (s Scripts) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(s) > 0 {
		scripts := xml.StartElement{
			Name: xml.Name{Local:"scripts"},
		}
		e.EncodeToken(scripts)
		e.Encode(([]*Script)(s))
		e.EncodeToken(scripts.End())
	}
	return nil
}

// UnmarshalXML correctly unmarshals a Structure and its children.
// This is needed because Go's xml parser doesn't handle the multiple unknown element, that
// need to be kept in order.
// TODO Do we need to keep things in order?
func (s *Structure) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {

	if err := unmarshalStartElement((*nakedStructure)(s), start); err != nil {
		return err
	}

	// Read the sub-elements and decode the correct ones
	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		switch token := token.(type) {
		case xml.StartElement:

			if token.Name.Local == "description" {
				// Special case the only non-element child
				if err = decoder.DecodeElement(&s.Description, &token); err != nil {
					return err
				}

			} else {
				var element Element

				switch token.Name.Local {
				// Elements:
				case "binary":
					element = &Binary{}
				case "custom":
					element = &Custom{}
				case "grammarref":
					element = &GrammarRef{}
				case "number":
					element = &Number{}
				case "offset":
					element = &Offset{}
				case "scriptelement":
					element = &ScriptElement{}
				case "string":
					element = &String{}
				case "structure":
					element = &Structure{}
				case "structref":
					element = &StructRef{}
				default:
					return fmt.Errorf("unknown element: `%s` inside `%s` at %d",
						token.Name.Local, start.Name.Local, decoder.InputOffset())
				}

				if err = decoder.DecodeElement(element, &token); err != nil {
					return err
				}

				s.Elements = append(s.Elements, element)
			}

		case xml.EndElement:
			return nil
		}
	}
}

/*
func marshalFixedValues(encoder *xml.Encoder, values []FixedValue) (error) {
	if len(values) == 0 {
		return nil
	}
	if len(values) == 1 {
		return encoder.Encode(values[0])
	}

	// TODO Check for returned error
	start := xml.StartElement{Name: xml.Name{Local: "fixedvalues"}}
	encoder.EncodeToken(start)
	for _, value := range values {
		encoder.Encode(value)
	}
	return encoder.EncodeToken(start.End())
}
*/

func (s *String) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	if err := unmarshalStartElement((*nakedString)(s), start); err != nil {
		return err
	}

	return s.unmarshalFixedValues(decoder, start)
}

func (s *String) unmarshalFixedValues(decoder *xml.Decoder, start xml.StartElement) error {

	for _, attr := range start.Attr {
		if attr.Name.Local == "fixedval" {
			s.Values = append(s.Values, &FixedValue{
				XMLName: xml.Name{Local: "fixedvalue"},
				Value:   attr.Value,
			})
		}
	}

	// We now read for all fixedvalues and fixedvalue children. We don't require
	// the fixedvalue to be inside a fixedvalues element.
	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		switch token := token.(type) {
		case xml.StartElement:
			var element interface{}

			switch token.Name.Local {
			case "fixedvalue":
				element = &FixedValue{}
			case "fixedvalues":
				element = &FixedValues{}
			case "description":
				element = &s.Description
			default:
				return fmt.Errorf("unknown element: `%s` inside `%s` at %d",
					token.Name.Local, start.Name.Local, decoder.InputOffset())
			}

			if err = decoder.DecodeElement(element, &token); err != nil {
				return err
			}

			switch token.Name.Local {
			case "fixedvalue":
				s.Values = append(s.Values, element.(*FixedValue))
			case "fixedvalues":
				s.Values = append(s.Values, element.(*FixedValues).Values...)
			}

		case xml.EndElement:
			return nil
		}
	}
}

func (b *Binary) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	if err := unmarshalStartElement((*nakedBinary)(b), start); err != nil {
		return err
	}

	return b.unmarshalFixedValues(decoder, start)
}

func (b *Binary) unmarshalFixedValues(decoder *xml.Decoder, start xml.StartElement) error {

	for _, attr := range start.Attr {
		if attr.Name.Local == "fixedval" {
			b.Values = append(b.Values, &FixedValue{
				XMLName: xml.Name{Local: "fixedvalue"},
				Value:   attr.Value,
			})
		}
	}

	// We now read for all fixedvalues and fixedvalue children. We don't require
	// the fixedvalue to be inside a fixedvalues element.
	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		switch token := token.(type) {
		case xml.StartElement:
			var element interface{}

			switch token.Name.Local {
			case "fixedvalue":
				element = &FixedValue{}
			case "fixedvalues":
				element = &FixedValues{}
			case "description":
				element = &b.Description
			default:
				return fmt.Errorf("unknown element: `%s` inside `%s` at %d",
					token.Name.Local, start.Name.Local, decoder.InputOffset())
			}

			if err = decoder.DecodeElement(element, &token); err != nil {
				return err
			}

			switch token.Name.Local {
			case "fixedvalue":
				b.Values = append(b.Values, element.(*FixedValue))
			case "fixedvalues":
				b.Values = append(b.Values, element.(*FixedValues).Values...)
			}

		case xml.EndElement:
			return nil
		}
	}
}

func (n *Number) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {

	if err := unmarshalStartElement((*nakedNumber)(n), start); err != nil {
		return err
	}

	return n.unmarshalFixedValues(decoder, start)
}

func (n *Number) unmarshalFixedValues(decoder *xml.Decoder, start xml.StartElement) error {

	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		switch token := token.(type) {
		case xml.StartElement:
			var element interface{}

			switch token.Name.Local {
			case "fixedvalue":
				element = &FixedValue{}
			case "fixedvalues":
				element = &FixedValues{}
			case "mask":
				element = &Mask{}
			case "description":
				element = &n.Description
			default:
				return fmt.Errorf("unknown element: `%s` inside `%s` at %d",
					token.Name.Local, start.Name.Local, decoder.InputOffset())
			}

			if err = decoder.DecodeElement(element, &token); err != nil {
				return err
			}

			switch token.Name.Local {
			case "fixedvalue":
				n.Values = append(n.Values, element.(*FixedValue))
			case "fixedvalues":
				n.Values = append(n.Values, element.(*FixedValues).Values...)
			case "mask":
				n.Masks = append(n.Masks, element.(*Mask))
			}

		case xml.EndElement:
			return nil
		}
	}
}

/*
func (s *String) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	// Replace start with a normally marshalled token
	start, err := marshalStartElement((*nakedString)(s))
	if err != nil {
		return err
	}

	// TODO Check for returned error
	encoder.EncodeToken(start)
	marshalFixedValues(encoder, s.Values)
	return encoder.EncodeToken(start.End())
}
*/
