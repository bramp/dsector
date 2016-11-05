package ufwb

import (
	"bramp.net/dsector/toerr"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	colourRegex = regexp.MustCompile("^[0-9A-F]{6}$")
)

// yesno returns the boolean value of this "yes", "no" field.
func yesno(s string, errs *toerr.Errors) Bool {
	// TODO Be strict on "yes", "no", "" only
	if s == "" {
		return UnknownBool
	}
	return boolOf(s == "yes")
}

// byteOrder returns the binary.byteOrder for this string.
func endian(s string, errs *toerr.Errors) Endian {
	switch s {
	case "big":
		return BigEndian
	case "little":
		return LittleEndian
	case "dynamic":
		return DynamicEndian
	case "":
		return UnknownEndian
	}

	errs.Append(fmt.Errorf("unknown endian: %q", s))
	return UnknownEndian
}

func display(s string, errs *toerr.Errors) Display {
	switch s {
	case "decimal":
		return DecDisplay
	case "hex":
		return HexDisplay
	case "binary":
		return BinaryDisplay
	case "":
		return UnknownDisplay
	}

	errs.Append(fmt.Errorf("unknown display: %q", s))
	return UnknownDisplay
}

func lengthunit(s string, errs *toerr.Errors) LengthUnit {
	switch s {
	case "bit":
		return BitLengthUnit
	case "byte":
		return ByteLengthUnit
	case "":
		return UnknownLengthUnit
	}

	errs.Append(fmt.Errorf("unknown length unit: %q", s))
	return UnknownLengthUnit
}

func colour(s string, errs *toerr.Errors) *Colour {
	if s == "" {
		return nil
	}

	if !colourRegex.MatchString(s) {
		errs.Append(fmt.Errorf("invalid colour: %q", s))
		return nil
	}

	b, err := hex.DecodeString(s)
	if err != nil {
		errs.Append(err)
		return nil
	}

	// TODO Check this is correct
	c := Colour(uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2]))
	return &c
}

func order(s string, errs *toerr.Errors) Order {
	switch s {
	case "fixed":
		return FixedOrder
	case "variable":
		return VariableOrder
	case "":
		return UnknownOrder
	}

	errs.Append(fmt.Errorf("unknown order: %q", s))
	return UnknownOrder
}

func (xml *XmlUfwb) transform() (*Ufwb, []error) {
	errs := &toerr.Errors{}

	u := &Ufwb{
		Xml:      xml,
		Version:  xml.Version,
		Elements: make(map[string]Element),
	}
	u.Grammar = xml.Grammar.transform(errs).(*Grammar)

	return u, errs.Slice()
}

func (xml *XmlGrammar) transform(errs *toerr.Errors) Element {
	g := &Grammar{
		Xml:      xml,
		Base:     xml.toBase("Grammar", errs),
		Author:   xml.Author,
		Ext:      xml.Ext,
		Email:    xml.Email,
		Complete: yesno(xml.Complete, errs),
		Uti:      xml.Uti,
	}

	if g.Xml.Start == "" {
		errs.Append(&validationError{e: g, err: errors.New("missing start attribute")})
	}

	for _, s := range xml.Scripts {
		g.Scripts = append(g.Scripts, s.transform(errs))
		// TODO Put the scripts into the Elements
	}

	for _, s := range xml.Structures {
		g.Elements = append(g.Elements, s.transform(errs))
	}

	return g
}

func (xml *XmlGrammarRef) transform(errs *toerr.Errors) Element {
	return &GrammarRef{
		Xml:  xml,
		Base: xml.toBase("GrammarRef", errs),

		uti:      xml.Uti,
		filename: xml.Filename,
		disabled: yesno(xml.Disabled, errs),
	}
}

func (xml *XmlStructure) transform(errs *toerr.Errors) Element {
	s := &Structure{
		Xml:  xml,
		Base: xml.toBase("Structure", errs),

		length:       Reference(xml.Length),
		lengthOffset: Reference(xml.LengthOffset),
		lengthUnit:   lengthunit(xml.LengthUnit, errs),

		Repeats: xml.toRepeats(errs),

		endian: endian(xml.Endian, errs),
		signed: yesno(xml.Signed, errs),

		encoding: xml.Encoding, // TODO Validate

		order: order(xml.Order, errs),

		display: display(xml.Display, errs),

		Colourful: Colourful{
			fillColour:   colour(xml.FillColour, errs),
			strokeColour: colour(xml.StrokeColour, errs),
		},
	}

	for _, e := range xml.Elements {
		s.elements = append(s.elements, e.transform(errs))
	}

	return s
}

func (xml *XmlCustom) transform(errs *toerr.Errors) Element {
	return &Custom{
		Xml:  xml,
		Base: xml.toBase("Custom", errs),

		length:     Reference(xml.Length),
		lengthUnit: lengthunit(xml.LengthUnit, errs),

		Colourful: Colourful{
			fillColour:   colour(xml.FillColour, errs),
			strokeColour: colour(xml.StrokeColour, errs),
		},
	}
}

func (xml *XmlStructRef) transform(errs *toerr.Errors) Element {
	return &StructRef{
		Xml:  xml,
		Base: xml.toBase("StructRef", errs),

		Repeats: xml.toRepeats(errs),

		disabled: yesno(xml.Disabled, errs),

		Colourful: Colourful{
			fillColour:   colour(xml.FillColour, errs),
			strokeColour: colour(xml.StrokeColour, errs),
		},
	}
}

// Parses a delimiter and returns the byte it represents. Currently the delimiter is required to
// be exact two hex characters, representing a single byte.
func delimiterToByte(delimiter string, errs *toerr.Errors) byte {
	if delimiter == "" {
		return 0
	}

	b, err := strconv.ParseUint(delimiter, 16, 8)
	if err != nil {
		errs.Append(fmt.Errorf("invalid delimiter %q: %s", delimiter, err))
	}
	return byte(b)
}

func (xml *XmlString) transform(errs *toerr.Errors) Element {
	s := &String{
		Xml:  xml,
		Base: xml.toBase("String", errs),

		typ:        xml.Type, // TODO Convert to "StringType" // "zero-terminated", "fixed-length", "pascal", "delimiter-terminated"
		length:     Reference(xml.Length),
		lengthUnit: lengthunit(xml.LengthUnit, errs),

		encoding:  xml.Encoding,
		mustMatch: yesno(xml.MustMatch, errs),

		delimiter: delimiterToByte(xml.Delimiter, errs),

		Repeats: xml.toRepeats(errs),

		Colourful: Colourful{
			fillColour:   colour(xml.FillColour, errs),
			strokeColour: colour(xml.StrokeColour, errs),
		},
	}

	if s.typ == "zero-terminated" {
		s.delimiter = '\x00'
	}

	for _, x := range xml.Values {
		s.values = append(s.values, &FixedStringValue{
			Xml:         x,
			name:        x.Name,
			value:       x.Value,
			description: strings.TrimSpace(x.Description),
		})
	}

	return s
}

func (xml *XmlBinary) transform(errs *toerr.Errors) Element {
	b := &Binary{
		Xml:  xml,
		Base: xml.toBase("Binary", errs),

		length:     Reference(xml.Length),
		lengthUnit: lengthunit(xml.LengthUnit, errs),

		Repeats: xml.toRepeats(errs),

		Colourful: Colourful{
			fillColour:   colour(xml.FillColour, errs),
			strokeColour: colour(xml.StrokeColour, errs),
		},

		mustMatch: yesno(xml.MustMatch, errs),
	}

	for _, x := range xml.Values {
		// Binary values shouldn't be prefixed, but incase they are:
		bs := strings.TrimPrefix(strings.TrimPrefix(x.Value, "0x"), "0X")
		value, err := hex.DecodeString(bs)
		if err != nil {
			errs.Append(err)
		}

		b.values = append(b.values, &FixedBinaryValue{
			Xml:         x,
			name:        x.Name,
			value:       value,
			description: strings.TrimSpace(x.Description),
		})
	}

	return b
}

func (xml *XmlNumber) transform(errs *toerr.Errors) Element {
	n := &Number{
		Xml:  xml,
		Base: xml.toBase("Number", errs),

		Type: xml.Type, // TODO Convert to NumberType

		length:     Reference(xml.Length),
		lengthUnit: lengthunit(xml.LengthUnit, errs),

		Repeats: xml.toRepeats(errs),

		endian: endian(xml.Endian, errs),
		signed: yesno(xml.Signed, errs),

		display: display(xml.Display, errs),

		Colourful: Colourful{
			fillColour:   colour(xml.FillColour, errs),
			strokeColour: colour(xml.StrokeColour, errs),
		},

		mustMatch: yesno(xml.MustMatch, errs),
	}

	for _, x := range xml.Values {
		n.values = append(n.values, &FixedValue{
			Xml:         x,
			name:        x.Name,
			description: strings.TrimSpace(x.Description),
		})
	}

	for _, v := range xml.Masks {
		n.masks = append(n.masks, v.transform(errs))
	}

	return n
}

func (xml *XmlOffset) transform(errs *toerr.Errors) Element {
	return &Offset{
		Xml:  xml,
		Base: xml.toBase("Offset", errs),

		length:     Reference(xml.Length),
		lengthUnit: lengthunit(xml.LengthUnit, errs),

		Repeats: xml.toRepeats(errs),

		endian: endian(xml.Endian, errs),

		display: display(xml.Display, errs),

		followNullReference: yesno(xml.FollowNullReference, errs),
		additional:          xml.Additional, // TODO Validate
	}
}

func (xml *XmlScriptElement) transform(errs *toerr.Errors) Element {
	return &ScriptElement{
		Xml:  xml,
		Base: xml.toBase("ScriptElement", errs),
	}
}

func (xml *XmlMask) transform(errs *toerr.Errors) *Mask {
	m := &Mask{
		Xml:         xml,
		name:        xml.Name,
		description: strings.TrimSpace(xml.Description),
	}

	for _, x := range xml.Values {
		// TODO Do I need to change this to some other type?
		m.values = append(m.values, &FixedValue{
			Xml:         x,
			name:        x.Name,
			description: strings.TrimSpace(x.Description),
		})
	}

	return m
}

func (xml *XmlScript) transform(errs *toerr.Errors) *Script {
	s := &Script{
		Xml:  xml,
		Name: xml.Name,
		Type: "Script",
	}

	if xml.Source != nil {
		s.Text = xml.Source.Text
		s.Language = xml.Source.Language
	} else {
		s.Text = xml.Text
	}

	if s.Language == "" {
		s.Language = xml.Language
	}

	return s
}
