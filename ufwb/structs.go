//go:ignore generate stringer -type Endian,Display,LengthUnit,Order
//go:generate getter -type Grammar,GrammarRef,Custom,String,Structure,StructRef,Binary,Number,Offset,ScriptElement,FixedValue
// TODO Consider moving this into a seperate package, so that the parser can't use the unexported fields (and forced to go via Getters, which "do the right thing" wrt extending and defaults.
package ufwb

import (
	"fmt"
	"io"
	"bramp.net/dsector/toerr"
)

const (
	Black = Colour(0x000000)
	White = Colour(0xffffff)
)

type Colour uint32
type Reference string // TODO At "parse time" check if this is a constant and record that.
type Bool int8        // tri-state bool unset, false, true.

// No other value is allowed
const (
	UnknownBool Bool = iota
	False
	True
)

func (b Bool) bool() bool {
	switch b {
	case False:
		return false
	case True:
		return true
	}
	panic("Unknown bool state")
}

func boolOf(b bool) Bool {
	if b {
		return True
	}
	return False
}

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
	panic(fmt.Sprintf("unknown base %s", d))
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
	FixedOrder         // TODO Check this is the right name
	VariableOrder
)

type Reader interface {
	// Read from file and return a Value.
	// The Read method must leave the file offset at Value.Offset + Value.Len // TODO Enforce this!
	Read(decoder *Decoder) (*Value, error)
}

type Formatter interface {
	// Format returns the display string for this Element.
	Format(file io.ReaderAt, value *Value) (string, error)
}

type Updatable interface {
	// Updates/validates the Element
	update(u *Ufwb, parent *Structure, errs *toerr.Errors)

	SetExtends(parent Element) error
}

type Repeatable interface {
	RepeatMin() Reference
	RepeatMax() Reference
}

// ElementLite is a light weight Element, one that only has a ID and a formatter.
type ElementLite interface {
	Id() int
	Name() string
	Description() string

	IdString() string

	Formatter
}

type Lengthable interface {
	Length() Reference
	LengthUnit() LengthUnit
}

type Element interface {
	ElementLite

	Reader
	Lengthable
	Repeatable
	Updatable

	// TODO Add Colourful here
}

type Colourful struct {
	fillColour   *Colour `default:"White" dereference:"true" parent:"false"`
	strokeColour *Colour `default:"Black" dereference:"true" parent:"false"`
}

type Ufwb struct {
	Xml *XmlUfwb

	Version string
	Grammar *Grammar

	Elements map[string]Element
}

// Base is what all Elements implement
type Base struct {
	elemType    string `setter:"false"` // This field is only for debug printing
	id          int
	name        string
	description string
}

func (b *Base) Id() int {
	return b.id
}
func (b *Base) Name() string {
	return b.name
}
func (b *Base) Description() string {
	return b.description
}

func (b *Base) GetBase() *Base {
	return b
}

func (b *Base) IdString() string {
	if b == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s<%02d %s>", b.elemType, b.id, b.name)
}

type Repeats struct {
	repeatMin Reference `default:"Reference(\"1\")" parent:"false"`
	repeatMax Reference `default:"Reference(\"1\")" parent:"false"`
}

type Grammar struct {
	Xml *XmlGrammar

	Base
	Repeats

	Author   string
	Ext      string
	Email    string
	Complete Bool
	Uti      string

	Start    Element // TODO Is this always a Structure?
	Elements []Element
	Scripts  []Script
}

type Structure struct {
	Xml *XmlStructure

	Base
	Repeats
	Colourful

	extends *Structure
	parent  *Structure

	length       Reference  `parent:"false"`
	lengthUnit   LengthUnit `default:"ByteLengthUnit"`
	lengthOffset Reference

	endian   Endian `default:"LittleEndian"`
	signed   Bool   `default:"True"`
	encoding string `default:"\"UTF-8\""`

	order Order `default:"FixedOrder"`

	display Display `default:"DecDisplay"`

	elements []Element `parent:"false"`

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
}

type GrammarRef struct {
	Xml *XmlGrammarRef

	Base
	Repeats

	extends *GrammarRef

	filename string
	grammar  *Grammar // TODO Actually implement this!

	//disabled bool
}

type Custom struct {
	Xml *XmlCustom

	Base
	Repeats
	Colourful

	extends *Custom

	length     Reference  `parent:"false"`
	lengthUnit LengthUnit `default:"ByteLengthUnit"`

	script string // TODO?
}

type StructRef struct {
	Xml *XmlStructRef

	Base
	Repeats
	Colourful

	extends *StructRef

	structure *Structure
}

type String struct {
	Xml *XmlString

	Base
	Repeats
	Colourful

	extends *String
	parent  *Structure

	typ string // TODO Convert to "StringType" // "zero-terminated", "fixed-length"

	length     Reference  `parent:"false"`
	lengthUnit LengthUnit `default:"ByteLengthUnit"`

	encoding string `default:"\"UTF-8\""`

	mustMatch Bool `default:"True"`
	values    []*FixedStringValue
}

type Binary struct {
	Xml *XmlBinary

	Base
	Repeats
	Colourful

	extends *Binary
	parent  *Structure

	length     Reference  `parent:"false"`
	lengthUnit LengthUnit `default:"ByteLengthUnit"`

	//unused     Bool // TODO
	//disabled   Bool

	mustMatch Bool `default:"True"`
	values    []*FixedBinaryValue
}

type Number struct {
	Xml *XmlNumber

	Base
	Repeats
	Colourful

	extends *Number
	parent  *Structure

	Type       string     // TODO Convert to Type
	length     Reference  `parent:"false"`
	lengthUnit LengthUnit `default:"ByteLengthUnit"`

	endian Endian `default:"LittleEndian"`
	signed Bool   `default:"True"`

	display Display `default:"DecDisplay"`

	// TODO Handle the below fields:
	valueExpression string

	minVal string // TODO This should be a int
	maxVal string

	mustMatch Bool `default:"True"`
	values    []*FixedValue
	masks     []*Mask
}

type Offset struct {
	Xml *XmlOffset

	Base
	Repeats
	Colourful

	extends *Offset

	length     Reference  `parent:"false"`
	lengthUnit LengthUnit `default:"ByteLengthUnit"`

	endian Endian `default:"LittleEndian"`

	RelativeTo          Element // TODO
	FollowNullReference string  // TODO
	References          Element // TODO
	ReferencedSize      Element // TODO
	Additional          string  // TODO

	display Display `default:"DecDisplay"`
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

	extends *FixedValue // TODO Can this actually be extended?

	// TODO Add description

	name  string
	value interface{}
}

type FixedBinaryValue struct {
	Xml *XmlFixedValue

	extends *FixedBinaryValue // TODO Can this actually be extended?

	name  string
	value []byte
}

type FixedStringValue struct {
	Xml *XmlFixedValue

	extends *FixedStringValue // TODO Can this actually be extended?

	name  string
	value string
}

// TODO Reconsider the script elements, as they don't need to correct match the XML

type ScriptElement struct {
	Xml *XmlScriptElement

	Base
	Repeats

	extends *ScriptElement

	//Disabled bool

	Script *Script
}

type Script struct {
	Xml *XmlScript

	Name string

	Type     string
	Language string

	Text string // TODO Sometimes there is a source element beneath this, pull it up into this field
}

type Padding struct {
	Base
}

type Elements []Element

func (e Elements) Find(name string) (int, Element) {
	for i, element := range e {
		if element.Name() == name {
			return i, element
		}
	}
	return -1, nil
}

// ElementByName returns a child element with this name, or nil
func (s *Structure) ElementByName(name string) Element {
	if _, e := Elements(s.elements).Find(name); e != nil {
		return e
	}

	if s.extends != nil {
		return s.extends.ElementByName(name)
	}

	return nil
}

func (s *Structure) Signed() bool {
	// TODO Move this to be auto generated
	if s.signed != UnknownBool {
		return s.signed.bool()
	}
	if s.extends != nil {
		return s.extends.Signed()
	}
	if s.parent != nil {
		return s.parent.Signed()
	}
	return true
}

func (s *Structure) SetSigned(signed bool) {
	s.signed = boolOf(signed)
}

func (n *Number) Signed() bool {
	if n.signed != UnknownBool {
		return n.signed.bool()
	}
	if n.extends != nil {
		return n.extends.Signed()
	}
	if n.parent != nil {
		return n.parent.Signed()
	}
	return true
}

func (n *Number) SetSigned(signed bool) {
	n.signed = boolOf(signed)
}

func (s *StructRef) Length() Reference {
	return s.Structure().Length()
}
func (s *StructRef) LengthUnit() LengthUnit {
	return s.Structure().LengthUnit()
}

func (g *GrammarRef) Length() Reference {
	return g.Grammar().Length()
}
func (g *GrammarRef) LengthUnit() LengthUnit {
	return g.Grammar().LengthUnit()
}

func (g *Grammar) Length() Reference {
	return "" // Always unset
}

func (g *Grammar) LengthUnit() LengthUnit {
	return ByteLengthUnit // Always unset
}

func (s *ScriptElement) Length() Reference {
	return "0" // ScriptElements have no form, thus no length
}

func (s *ScriptElement) LengthUnit() LengthUnit {
	return ByteLengthUnit
}
