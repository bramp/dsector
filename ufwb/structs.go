//go:ignore generate stringer -type Endian,Display,LengthUnit,Order
//go:generate getter -type Grammar,GrammarRef,Custom,String,Structure,StructRef,Binary,Number,Offset,ScriptElement,FixedValue
// TODO Consider moving this into a seperate package, so that the parser can't use the unexported fields (and forced to go via Getters, which "do the right thing" wrt extending and defaults.
package ufwb

import (
	"fmt"
	"strconv"
)

const (
	Black = Colour(0x000000)
	White = Colour(0xffffff)
)

type Colour uint32
type Reference string
type Bool int8 // tri-state bool unset, false, true.

// No other value is allowed
const (
	UnknownBool Bool = iota
	False
	True
)

func (b Bool) bool() bool {
	switch b {
	case False: return false
	case True: return true
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

	SetExtends(parent Element) error
}

type Repeatable interface {
	RepeatMin() Reference
	RepeatMax() Reference
}


type Element interface {
	Reader
	Formatter

	Repeatable
	Updatable

	Id() int
	Name() string
	Description() string

	String() string

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
	elemType    string
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

func (b *Base) String() string {
	return b.debugString()
}

func (b *Base) debugString() string {
	return fmt.Sprintf("%s<%02d %s>", b.elemType, b.id, b.name)
}

type Repeats struct {
	repeatMin    Reference  `default:"Reference(\"1\")"`
	repeatMax    Reference  `default:"Reference(\"1\")"`
}

/*
var (
	defaultGrammar = Grammar{}
	defaultStructure = Structure{
		endian:  LittleEndian, // TODO Check this is the right default
		signed:  True,         // TODO Check this is the right default
		display: DecDisplay, // TODO Check this is the right default
		lengthUnit: ByteLengthUnit, // TODO Check this is the right default
		encoding: "UTF-8", // TODO Check this is the right default
		order: FixedOrder, // TODO Check this is the right default
	}
	defaultGrammarRef = GrammarRef{}
	defaultCustom = Custom{}
	defaultStructRef = StructRef{}
	defaultString = String{}
	defaultBinary = Binary{}
	defaultNumber = Number{}
	defaultOffset = Offset{}
)
*/

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
	Xml          *XmlStructure

	Base
	Repeats
	Colourful

	extends      *Structure
	parent       *Structure

	length       Reference
	lengthUnit   LengthUnit `default:"ByteLengthUnit"`
	lengthOffset Reference

	endian       Endian `default:"LittleEndian"`
	signed       Bool   `default:"True"`
	encoding     string `default:"\"UTF-8\""`

	order        Order `default:"FixedOrder"`

	display      Display `default:"DecDisplay"`

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
}

type GrammarRef struct {
	Xml *XmlGrammarRef

	Base
	Repeats

	extends      *GrammarRef

	filename string
	//disabled bool
}

type Custom struct {
	Xml *XmlCustom

	Base
	Repeats
	Colourful

	extends      *Custom

	length            Reference
	lengthUnit        LengthUnit `default:"ByteLengthUnit"`

	script string // TODO?
}

type StructRef struct {
	Xml *XmlStructRef

	Base
	Repeats
	Colourful

	extends   *StructRef

	structure *Structure
}

type String struct {
	Xml *XmlString

	Base
	Repeats
	Colourful

	extends      *String

	typ string // TODO Convert to "StringType" // "zero-terminated", "fixed-length"

	length        Reference
	lengthUnit    LengthUnit `default:"ByteLengthUnit"`

	encoding  string `default:"\"UTF-8\""`

	mustMatch Bool `default:"True"`
	values []*FixedStringValue
}

type Binary struct {
	Xml        *XmlBinary

	Base
	Repeats
	Colourful

	extends    *Binary
	parent     *Structure

	length     Reference
	lengthUnit LengthUnit `default:"ByteLengthUnit"`

	//unused     Bool // TODO
	//disabled   Bool

	mustMatch  Bool `default:"True"`
	values     []*FixedBinaryValue
}

type Number struct {
	Xml             *XmlNumber

	Base
	Repeats
	Colourful

	extends         *Number
	parent          *Structure

	Type            string // TODO Convert to Type
	length          Reference
	lengthUnit      LengthUnit `default:"ByteLengthUnit"`

	endian          Endian  `default:"LittleEndian"`
	signed          Bool    `default:"True"`

	display         Display `default:"DecDisplay"`

	// TODO Handle the below fields:
	valueExpression string

	minVal          string // TODO This should be a int
	maxVal          string

	mustMatch       Bool   `default:"True"`
	values          []*FixedValue
	masks           []*Mask
}

// Bytes returns the width of this number in bytes
func (n *Number) Bytes() int {
	// TODO Change this to use Eval
	len, err := strconv.Atoi(string(n.Length()))
	if err != nil {
		panic("TODO USE EVAL")
	}

	if n.LengthUnit() == BitLengthUnit {
		return len / 8
	}
	if n.LengthUnit() == ByteLengthUnit {
		return len
	}

	panic("Unknown length unit")
}

func (n *Number) Bits() int {
	// TODO Change this to use Eval
	len, err := strconv.Atoi(string(n.Length()))
	if err != nil {
		panic("TODO USE EVAL")
	}

	if n.LengthUnit() == BitLengthUnit {
		return len
	}
	if n.LengthUnit() == ByteLengthUnit {
		return len * 8
	}

	panic("Unknown length unit")
}

type Offset struct {
	Xml *XmlOffset

	Base
	Repeats
	Colourful

	extends         *Offset

	length Reference
	lengthUnit      LengthUnit `default:"ByteLengthUnit"`

	endian Endian  `default:"LittleEndian"`

	RelativeTo          Element // TODO
	FollowNullReference string  // TODO
	References          Element // TODO
	ReferencedSize      Element // TODO
	Additional          string  // TODO

	display         Display `default:"DecDisplay"`

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

	// TODO Add description

	name  string
	value interface{}
}

type FixedBinaryValue struct {
	Xml *XmlFixedValue

	extends         *FixedBinaryValue // TODO Can this actually be extended?

	name  string
	value []byte
}

type FixedStringValue struct {
	Xml *XmlFixedValue

	extends         *FixedStringValue // TODO Can this actually be extended?

	name  string
	value string
}

// TODO Reconsider the script elements, as they don't need to correct match the XML

type ScriptElement struct {
	Xml *XmlScriptElement

	Base
	Repeats

	extends         *ScriptElement

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

func (s *Structure) Signed() bool {
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

