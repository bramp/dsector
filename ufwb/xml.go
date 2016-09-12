// Shout out to https://github.com/wicast/xj2s for helping to generate the XML structs
package ufwb

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

type Transformable interface {
	// transform creates a new naitve Element to represent this XMLElement.
	// Only the Base fields are transformed at this point, Name, ID, Description. The rest are validated
	// and parsed at a later stage, when more context is available.
	transform(errs *Errors) Element
}

type XmlElement interface {
	Transformable
}

type XmlIdName struct {
	Id          int    `xml:"id,attr,omitempty"`
	Name        string `xml:"name,attr,omitempty"`
	Description string `xml:"description,omitempty"`
}

func (xml *XmlIdName) toIdName(errs *Errors) Base {
	return Base{
		Id:          xml.Id,
		Name:        xml.Name,
		Description: xml.Description,
	}
}

type XmlUfwb struct {
	XMLName xml.Name `xml:"ufwb"`

	Version string      `xml:"version,attr,omitempty"`
	Grammar *XmlGrammar `xml:"grammar"`
}

func (xml *XmlUfwb) transform() (*Ufwb, []error) {
	errs := &Errors{}

	u := &Ufwb{
		Xml:     xml,
		Version: xml.Version,
		Elements: make(map[string]Element),
	}
	u.Grammar = xml.Grammar.transform(errs).(*Grammar)

	return u, errs.Slice()
}

type XmlGrammar struct {
	XMLName    xml.Name `xml:"grammar"`

	XmlIdName

	Author     string `xml:"author,attr,omitempty"`
	Ext        string `xml:"fileextension,attr,omitempty"`
	Email      string `xml:"email,attr,omitempty"`
	Complete   string `xml:"complete,attr,omitempty" ufwb:"bool"`
	Uti        string `xml:"uti,attr,omitempty"`

	Start      string          `xml:"start,attr,omitempty" ufwb:"id"`
	Scripts    XmlScripts      `xml:"scripts"`
	Structures []*XmlStructure `xml:"structure,omitempty"`
}

func (xml *XmlGrammar) transform(errs *Errors) Element {

	g := &Grammar{
		Xml:    xml,
		Base: xml.XmlIdName.toIdName(errs),
	}

	for _, s := range xml.Structures {
		g.Elements = append(g.Elements, s.transform(errs))
	}

	return g
}

type XmlGrammarRef struct {
	XMLName  xml.Name `xml:"grammarref"`

	XmlIdName
	Uti      string `xml:"uti,attr,omitempty"`

	Filename string `xml:"filename,attr,omitempty"`
	Disabled string `xml:"disabled,attr,omitempty" ufwb:"bool"`
}

func (xml *XmlGrammarRef) transform(errs *Errors) Element {
	return &GrammarRef{
		Xml:    xml,
		Base: xml.XmlIdName.toIdName(errs),
	}
}

type XmlStructure struct {
	XMLName         xml.Name `xml:"structure"`

	XmlIdName

	Length          string `xml:"length,attr,omitempty" ufwb:"ref"`
	LengthUnit      string `xml:"lengthunit,attr,omitempty" ufwb:"lengthunit"`

	LengthOffset    string `xml:"lengthoffset,attr,omitempty"`

	Endian          string `xml:"endian,attr,omitempty" ufwb:"endian"`
	Signed          string `xml:"signed,attr,omitempty" ufwb:"bool"`
	Extends         string `xml:"extends,attr,omitempty" ufwb:"id"`
	Order           string `xml:"order,attr,omitempty"`
	Encoding        string `xml:"encoding,attr,omitempty" ufwb:"encoding"`
	Alignment       string `xml:"alignment,attr,omitempty"` // ??

	Floating        string `xml:"floating,attr,omitempty"`  // ??
	ConsistsOf      string `xml:"consists-of,attr,omitempty" ufwb:"id"`

	Repeat          string `xml:"repeat,attr,omitempty" ufwb:"id"`
	RepeatMin       string `xml:"repeatmin,attr,omitempty" ufwb:"ref"`
	RepeatMax       string `xml:"repeatmax,attr,omitempty" ufwb:"ref"`

	ValueExpression string `xml:"valueexpression,attr,omitempty"`
	Debug           string `xml:"debug,attr,omitempty" ufwb:"bool"`
	Disabled        string `xml:"disabled,attr,omitempty" ufwb:"bool"`

	Display         string `xml:"display,attr,omitempty" ufwb:"display"`
	FillColour      string `xml:"fillcolor,attr,omitempty" ufwb:"colour"`
	StrokeColour    string `xml:"strokecolor,attr,omitempty" ufwb:"colour"`

	Elements        []XmlElement `xml:",any"`
}

func (xml *XmlStructure) transform(errs *Errors) Element {
	s := &Structure{
		Xml:    xml,
		Base: xml.XmlIdName.toIdName(errs),
	}

	for _, e := range xml.Elements {
		s.elements = append(s.elements, e.transform(errs))
	}

	return s
}

type XmlCustom struct {
	XMLName      xml.Name `xml:"custom"`

	XmlIdName

	Length       string `xml:"length,attr,omitempty" ufwb:"ref"`
	Script       string `xml:"script,attr,omitempty" ufwb:"id"`

	FillColour   string `xml:"fillcolor,attr,omitempty" ufwb:"colour"`
	StrokeColour string `xml:"strokecolor,attr,omitempty" ufwb:"colour"`
}

func (xml *XmlCustom) transform(errs *Errors) Element {
	return &Custom{
		Xml:    xml,
		Base: xml.XmlIdName.toIdName(errs),
	}
}

type XmlStructRef struct {
	XMLName      xml.Name `xml:"structref"`

	XmlIdName

	Structure    string `xml:"structure,attr,omitempty" ufwb:"id"`
	RepeatMin    string `xml:"repeatmin,attr,omitempty" ufwb:"ref"`
	RepeatMax    string `xml:"repeatmax,attr,omitempty" ufwb:"ref"`

	FillColour   string `xml:"fillcolor,attr,omitempty" ufwb:"colour"`
	StrokeColour string `xml:"strokecolor,attr,omitempty" ufwb:"colour"`
}

func (xml *XmlStructRef) transform(errs *Errors) Element {
	return &StructRef{
		Xml:    xml,
		Base: xml.XmlIdName.toIdName(errs),
	}
}

type XmlString struct {
	XMLName      xml.Name `xml:"string"`

	XmlIdName

	Type         string `xml:"type,attr,omitempty" ufwb:"string-type"`  // "zero-terminated", "fixed-length"

	Length       string `xml:"length,attr,omitempty" ufwb:"ref"`
	Encoding     string `xml:"encoding,attr,omitempty" ufwb:"encoding"` // Should be valid encoding
	MustMatch    string `xml:"mustmatch,attr,omitempty" ufwb:"bool"`    // "yes", "no"

	RepeatMin    string `xml:"repeatmin,attr,omitempty" ufwb:"ref"`
	RepeatMax    string `xml:"repeatmax,attr,omitempty" ufwb:"ref"`

	FillColour   string `xml:"fillcolor,attr,omitempty" ufwb:"colour"`
	StrokeColour string `xml:"strokecolor,attr,omitempty" ufwb:"colour"`

	Values       []*XmlFixedValue `xml:"fixedvalue,omitempty"`
}

func (xml *XmlString) transform(errs *Errors) Element {
	s := &String{
		Xml:    xml,
		Base: xml.XmlIdName.toIdName(errs),
	}

	for _, x := range xml.Values {
		s.values = append(s.values, &FixedStringValue{
			Xml:   x,
			name:  x.Name,
		})
	}

	return s
}

type XmlBinary struct {
	XMLName      xml.Name `xml:"binary"`

	XmlIdName

	Length       string `xml:"length,attr,omitempty" ufwb:"ref"`
	LengthUnit   string `xml:"lengthunit,attr,omitempty" ufwb:"lengthunit"` // "bit"

	RepeatMin    string `xml:"repeatmin,attr,omitempty" ufwb:"ref"`
	RepeatMax    string `xml:"repeatmax,attr,omitempty" ufwb:"ref"`

	MustMatch    string `xml:"mustmatch,attr,omitempty"  ufwb:"bool"`
	Unused       string `xml:"unused,attr,omitempty" ufwb:"bool"`
	Disabled     string `xml:"disabled,attr,omitempty" ufwb:"bool"`

	FillColour   string `xml:"fillcolor,attr,omitempty" ufwb:"colour"`
	StrokeColour string `xml:"strokecolor,attr,omitempty" ufwb:"colour"`

	Values       []*XmlFixedValue `xml:"fixedvalue,omitempty"`
}

func (xml *XmlBinary) transform(errs *Errors) Element {
	b := &Binary{
		Xml:    xml,
		Base: xml.XmlIdName.toIdName(errs),
	}

	for _, x := range xml.Values {
		b.values = append(b.values, &FixedBinaryValue{
			Xml:   x,
			name:  x.Name,
		})
	}

	return b
}

type XmlNumber struct {
	XMLName         xml.Name `xml:"number"`

	XmlIdName

	RepeatMin       string `xml:"repeatmin,attr,omitempty" ufwb:"ref"`
	RepeatMax       string `xml:"repeatmax,attr,omitempty" ufwb:"ref"`

	Type            string `xml:"type,attr,omitempty" ufwb:"number-type"`
	Length          string `xml:"length,attr,omitempty" ufwb:"ref"`
	LengthUnit      string `xml:"lengthunit,attr,omitempty" ufwb:"lengthunit"` // "", "bit" (default "byte")

	Endian          string `xml:"endian,attr,omitempty" ufwb:"endian"`         // "", "big", "little", "dynamic"
	Signed          string `xml:"signed,attr,omitempty" ufwb:"bool"`           // "", "yes", "no"
	MustMatch       string `xml:"mustmatch,attr,omitempty" ufwb:"bool"`
	ValueExpression string `xml:"valueexpression,attr,omitempty"`

	MinVal          string `xml:"minval,attr,omitempty" ufwb:"ref"`
	MaxVal          string `xml:"maxval,attr,omitempty" ufwb:"ref"`

	Display         string `xml:"display,attr,omitempty" ufwb:"display"`
	FillColour      string `xml:"fillcolor,attr,omitempty" ufwb:"colour"`
	StrokeColour    string `xml:"strokecolor,attr,omitempty" ufwb:"colour"`

	Disabled        string `xml:"disabled,attr,omitempty" ufwb:"bool"`

	Values          []*XmlFixedValue `xml:"fixedvalue,omitempty"`
	Masks           []*XmlMask       `xml:"mask,omitempty"`
}

func (xml *XmlNumber) transform(errs *Errors) Element {
	n := &Number{
		Xml:    xml,
		Base: xml.XmlIdName.toIdName(errs),
	}

	for _, x := range xml.Values {
		n.values = append(n.values, &FixedValue{
			Xml:   x,
			name:  x.Name,
		})
	}

	for _, v := range xml.Masks {
		n.masks = append(n.masks, v.transform(errs))
	}

	return n
}

type XmlOffset struct {
	XMLName             xml.Name `xml:"offset"`

	XmlIdName

	RepeatMin           string `xml:"repeatmin,attr,omitempty" ufwb:"ref"`
	RepeatMax           string `xml:"repeatmax,attr,omitempty" ufwb:"ref"`

	Length              string `xml:"length,attr,omitempty" ufwb:"ref"`
	Endian              string `xml:"endian,attr,omitempty" ufwb:"endian"`
	RelativeTo          string `xml:"relative-to,attr,omitempty" ufwb:"id"`
	FollowNullReference string `xml:"follownullreference,attr,omitempty"`
	References          string `xml:"references,attr,omitempty" ufwb:"id"`
	ReferencedSize      string `xml:"referenced-size,attr,omitempty" ufwb:"id"`
	Additional          string `xml:"additional,attr,omitempty"`             // "stringOffset"

	Display             string `xml:"display,attr,omitempty" ufwb:"display"` // "", "hex", "offset"
	FillColour          string `xml:"fillcolor,attr,omitempty" ufwb:"colour"`
	StrokeColour        string `xml:"strokecolor,attr,omitempty" ufwb:"colour"`
}

func (xml *XmlOffset) transform(errs *Errors) Element {
	return &Offset{
		Xml:    xml,
		Base: xml.XmlIdName.toIdName(errs),
	}
}

type XmlScriptElement struct {
	XMLName  xml.Name `xml:"scriptelement"`

	XmlIdName
	Disabled string `xml:"disabled,attr,omitempty" ufwb:"bool"`

	Script   *XmlScript `xml:"script"`
}

func (xml *XmlScriptElement) transform(errs *Errors) Element {
	return &ScriptElement{
		Xml:    xml,
		Base: xml.XmlIdName.toIdName(errs),
	}
}

type XmlMask struct {
	XMLName xml.Name `xml:"mask"`

	Name    string `xml:"name,attr,omitempty"`
	Value   string `xml:"value,attr,omitempty"`

	Values  []*XmlFixedValue `xml:"fixedvalue,omitempty"`
}

func (xml *XmlMask) transform(errs *Errors) *Mask {
	m := &Mask{
		Xml:  xml,
		name: xml.Name,
	}

	for _, x := range xml.Values {
		// TODO Do I need to change this to some other type?
		m.values = append(m.values, &FixedValue{
			Xml:   x,
			name:  x.Name,
		})
	}

	return m
}

type XmlScripts []*XmlScript

type XmlScript struct {
	XMLName  xml.Name `xml:"script"`

	XmlIdName

	Type     string     `xml:"type,attr,omitempty"` // DataType, Grammar
	Language string     `xml:"language,attr,omitempty" ufwb:"lang"`
	Source   *XmlSource `xml:"source,omitempty"`

													// Sometimes the text is defined here, or in the child Source element
	Text     string `xml:",chardata"`               // TODO Should this be cdata?
}

func (xml *XmlScript) transform(errs *Errors) *Script {
	return &Script{
		Xml:  xml,
		Name: xml.Name,
		// TODO Pull info from XmlSource tag
	}
}

type XmlSource struct {
	XMLName  xml.Name `xml:"source"`

	Language string `xml:"language,attr,omitempty" ufwb:"lang"`
	Text     string `xml:",chardata"` // TODO Should this be cdata?

	language string
}

type XmlFixedValues struct {
	XMLName xml.Name `xml:"fixedvalues"`

	Values  []*XmlFixedValue `xml:"fixedvalue,omitempty"`
}

type XmlFixedValue struct {
	XMLName xml.Name `xml:"fixedvalue"`

	Name    string `xml:"name,attr,omitempty"`
	Value   string `xml:"value,attr,omitempty"`
}

/*
func (xml *XmlFixedValue) transform(errs *Errors) *FixedValue {
	return &FixedValue{
		Xml:   xml,
		name:  xml.Name,
	}
}
*/

// Types of the original elements but without the MarshalXML / UnmarshalXML methods on them.
type nakedXmlStructure XmlStructure
type nakedXmlString XmlString
type nakedXmlNumber XmlNumber
type nakedXmlBinary XmlBinary

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

func (s *XmlScripts) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		switch token := token.(type) {
		case xml.StartElement:
			if token.Name.Local == "script" {
				element := &XmlScript{}
				if err = decoder.DecodeElement(element, &token); err != nil {
					return err
				}
				if element.Source != nil {
					// If there is a source element, then the Text shouldn't be set
					element.Text = ""
				}
				*s = append(([]*XmlScript)(*s), element)
			} else {
				return fmt.Errorf("unknown element: `%s` inside `%s` at %d",
					token.Name.Local, start.Name.Local, decoder.InputOffset())
			}
		case xml.EndElement:
			return nil
		}
	}
}

func (s XmlScripts) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(s) > 0 {
		scripts := xml.StartElement{
			Name: xml.Name{Local: "scripts"},
		}
		e.EncodeToken(scripts)
		e.Encode(([]*XmlScript)(s))
		e.EncodeToken(scripts.End())
	}
	return nil
}

// UnmarshalXML correctly unmarshals a Structure and its children.
// This is needed because Go's xml parser doesn't handle the multiple unknown element, that
// need to be kept in order.
// TODO Do we need to keep things in order?
func (s *XmlStructure) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {

	if err := unmarshalStartElement((*nakedXmlStructure)(s), start); err != nil {
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
				var element XmlElement

				switch token.Name.Local {
				// Elements:
				case "binary":
					element = &XmlBinary{}
				case "custom":
					element = &XmlCustom{}
				case "grammarref":
					element = &XmlGrammarRef{}
				case "number":
					element = &XmlNumber{}
				case "offset":
					element = &XmlOffset{}
				case "scriptelement":
					element = &XmlScriptElement{}
				case "string":
					element = &XmlString{}
				case "structure":
					element = &XmlStructure{}
				case "structref":
					element = &XmlStructRef{}
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

func (s *XmlString) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	if err := unmarshalStartElement((*nakedXmlString)(s), start); err != nil {
		return err
	}

	return s.unmarshalFixedValues(decoder, start)
}

func (s *XmlString) unmarshalFixedValues(decoder *xml.Decoder, start xml.StartElement) error {

	for _, attr := range start.Attr {
		if attr.Name.Local == "fixedval" {
			s.Values = append(s.Values, &XmlFixedValue{
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
				element = &XmlFixedValue{}
			case "fixedvalues":
				element = &XmlFixedValues{}
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
				s.Values = append(s.Values, element.(*XmlFixedValue))
			case "fixedvalues":
				s.Values = append(s.Values, element.(*XmlFixedValues).Values...)
			}

		case xml.EndElement:
			return nil
		}
	}
}

func (b *XmlBinary) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	if err := unmarshalStartElement((*nakedXmlBinary)(b), start); err != nil {
		return err
	}

	return b.unmarshalFixedValues(decoder, start)
}

func (b *XmlBinary) unmarshalFixedValues(decoder *xml.Decoder, start xml.StartElement) error {

	for _, attr := range start.Attr {
		if attr.Name.Local == "fixedval" {
			b.Values = append(b.Values, &XmlFixedValue{
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
				element = &XmlFixedValue{}
			case "fixedvalues":
				element = &XmlFixedValues{}
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
				b.Values = append(b.Values, element.(*XmlFixedValue))
			case "fixedvalues":
				b.Values = append(b.Values, element.(*XmlFixedValues).Values...)
			}

		case xml.EndElement:
			return nil
		}
	}
}

func (n *XmlNumber) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {

	if err := unmarshalStartElement((*nakedXmlNumber)(n), start); err != nil {
		return err
	}

	return n.unmarshalFixedValues(decoder, start)
}

func (n *XmlNumber) unmarshalFixedValues(decoder *xml.Decoder, start xml.StartElement) error {

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
				element = &XmlFixedValue{}
			case "fixedvalues":
				element = &XmlFixedValues{}
			case "mask":
				element = &XmlMask{}
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
				n.Values = append(n.Values, element.(*XmlFixedValue))
			case "fixedvalues":
				n.Values = append(n.Values, element.(*XmlFixedValues).Values...)
			case "mask":
				n.Masks = append(n.Masks, element.(*XmlMask))
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
