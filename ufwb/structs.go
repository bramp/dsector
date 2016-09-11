//go:ignore generate stringer -type Endian,Display,LengthUnit,Order
//go:generate getter -type Grammar,GrammarRef,Custom,String,Structure,StructRef,Binary,Number,Offset,ScriptElement,FixedValue
package ufwb

import (
	"fmt"
	"strconv"
)

type Colour uint32
type Reference string

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
	case HexDisplay:
		return 16
	case DecDisplay:
		return 10
	case BinaryDisplay:
		return 2
	case UnknownDisplay:
		return 0
	}
	return 0
}

type LengthUnit int

const (
	UnknownLengthUnit LengthUnit = iota
	BitLengthUnit
	ByteLengthUnit
)

type Order int

const (
	UnknownOrder Order = iota
	FixedOrder // TODO Check this is the right name
	VariableOrder
)


type Reader interface {
	// Read from file and return a Value.
	// The Read method must leave the file offset at Value.Offset + Value.Len
	Read(decoder *Decoder) (*Value, error)
}

type Formatter interface {
	// Format returns the display string for this Element
	Format(file File, value *Value) (string, error)
}

type Updatable interface {
	// Updates/validates the Element
	update(u *Ufwb, parent *Structure, errs *Errors)

	// Checks if we extend. This is done before update() because extending may impact
	// our future parsing.
	updateExtends(u *Ufwb) error
}


type Element interface {
	Reader
	Formatter

	Updatable

	// Rename these to just Id(), Name(), Description()
	GetId() int
	GetName() string
	GetDescription() string

	// TODO Add Colourful here

}

type Colourful struct {
	fillColour   Colour
	strokeColour Colour
}

type Ufwb struct {
	Xml *XmlUfwb

	Version string
	Grammar *Grammar

	Elements map[string]Element
}

// Base is what all Elements implement
type Base struct {
	Id          int
	Name        string
	Description string
}

func (b *Base) GetId() int {
	return b.Id
}
func (b *Base) GetName() string {
	return b.Name
}
func (b *Base) GetDescription() string {
	return b.Description
}

func (b *Base) String() string {
	return fmt.Sprintf("<%02d %s>", b.Id, b.Name)
}

type Grammar struct {
	Xml *XmlGrammar

	Base

	Author   string
	Ext      string
	Email    string
	Complete bool
	Uti      string

	Start    Element // TODO Is this always a Structure?
	Elements []Element
	Scripts  []Script
}


type Structure struct {
	Xml          *XmlStructure

	Base
	Colourful
	extends      *Structure

	length       Reference
	lengthUnit   LengthUnit
	lengthOffset Reference

	endian       Endian
	signed       bool
	encoding     string

	order        Order

	//Display   Display

	elements     []Element

	/*
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
	*/

	display      Display
}

func (s *Structure) Length() Reference {
	// TODO THis is a test
	if s.length == "" {
		return s.extends.Length()
	}
	return s.length
}


func (s *Structure) SetExtends(element Element) (error) {
	parent, ok := element.(*Structure);
	if !ok {
		return &validationError{e: s, msg: fmt.Sprintf("element can't extend from %T", element)}
	}
	s.extends = parent

	if len(s.elements) < len(parent.elements) {
		return &validationError{e: s, msg: "child element must have atleast as many elements as the parent"}
	}

	/*
	for i, child := range parent.elements {
		if e, ok := s.elements[i].(Updatable); ok {
			err := e.SetExtends(child)
			if err != nil {
				return err
			}
		}
	}
	*/

	return nil
}

type GrammarRef struct {
	Xml *XmlGrammarRef

	Base
	extends      *GrammarRef

	filename string
	disabled bool
}

type Custom struct {
	Xml *XmlCustom

	Base
	Colourful
	extends      *Custom

	length Reference
	script string // TODO?
}

type StructRef struct {
	Xml *XmlStructRef

	Base
	Colourful
	extends   *StructRef

	structure *Structure

	repeatMin Reference
	repeatMax Reference
}

type String struct {
	Xml *XmlString

	Base
	Colourful
	extends      *String

	typ string // TODO Convert to "StringType" // "zero-terminated", "fixed-length"

	length    Reference
	encoding  string
	mustMatch bool

	repeatMin Reference
	repeatMax Reference

	values []*FixedValue
}

type Binary struct {
	Xml        *XmlBinary

	Base
	Colourful
	extends    *Binary

	length     Reference
	lengthUnit LengthUnit

	repeatMin  Reference
	repeatMax  Reference

	mustMatch  bool
	unused     bool
	disabled   bool

	values     []*FixedValue
}

func (b *Binary) Length() Reference {
	// TODO THis is a test
	if b.length == "" {
		return b.extends.Length()
	}
	return b.length
}

func (b *Binary) SetExtends(element Element) (error) {
	parent, ok := element.(*Binary);
	if !ok {
		return &validationError{e: b, msg: fmt.Sprintf("element can't extend from %T", element)}
	}
	b.extends = parent
	return nil
}

type Number struct {
	Xml             *XmlNumber

	Base
	Colourful
	extends         *Number

	repeatMin       Reference
	repeatMax       Reference

	Type            string // TODO Convert to Type
	length          Reference
	lengthUnit      LengthUnit

	endian          Endian
	signed          bool

	display         Display

	// TODO Handle the below fields:
	mustMatch       bool
	valueExpression string

	minVal          string
	maxVal          string

	values          []*FixedValue
	masks           []*Mask
}

func (n *Number) Bits() int {
	// TODO Change this to use Eval
	len, err := strconv.Atoi(string(n.Length()))
	if err != nil {
		return -1
	}

	if n.LengthUnit() == BitLengthUnit {
		return len
	}
	if n.LengthUnit() == ByteLengthUnit {
		return len * 8
	}

	return -1
}

func (s *Number) SetExtends(element Element) (error) {
	parent, ok := element.(*Number);
	if !ok {
		return &validationError{e: s, msg: fmt.Sprintf("element can't extend from %T", element)}
	}
	s.extends = parent
	return nil
}

type Offset struct {
	Xml *XmlOffset

	Base
	extends         *Number

	RepeatMin Reference
	RepeatMax Reference

	Length Reference
	Endian Endian

	RelativeTo          Element `xml:"relative-to,attr,omitempty" ufwb:"id"`
	FollowNullReference string  `xml:"follownullreference,attr,omitempty"` // TODO
	References          Element `xml:"references,attr,omitempty" ufwb:"id"`
	ReferencedSize      Element `xml:"referenced-size,attr,omitempty" ufwb:"id"`
	Additional          string  `xml:"additional,attr,omitempty"` // "stringOffset" // TODO

	Display Display

	Colourful
}

type Mask struct {
	Xml *XmlMask

	name  string
	value uint64 // The mask

	values []*FixedValue
}

/*
type FixedValues struct {
	Xml *XmlFixedValues

	values []*FixedValue
}
*/

type FixedValue struct {
	Xml *XmlFixedValue

	extends         *FixedValue // TODO Can this actually be extended?

	name  string
	value interface{}
}

// TODO Reconsider the script elements, as they don't need to correct match the XML

type ScriptElement struct {
	Xml *XmlScriptElement

	Base

	Disabled bool

	Script *Script
}

type Script struct {
	Xml *XmlScript

	Name string

	Type     string
	Language string

	Text string // TODO Sometimes there is a source element beneath this, pull it up into this field
}
