package ufwb

import (
	"fmt"
)

type Reference string

type Reader interface {
	// Read from file and return a Value.
	// The Read method must leave the file offset at Value.Offset + Value.Len
	Read(decoder *Decoder) (*Value, error)
}

type Formatter interface {
	// Format returns the display string for this Element
	Format(file File, value *Value) (string, error)
}

type Element interface {
	Reader
	Formatter

	GetId() int
	GetName() string
	GetDescription() string

	// Updates/validates the Element
	update(u *Ufwb, parent *Structure, errs []error)
}

type Ufwb struct {
	Xml *XmlUfwb

	Version string
	Grammar *Grammar

	Elements map[string]Element
}

func (u *Ufwb) Get(id string) (*Structure, error) {
	element, found := u.Elements[id]
	if !found {
		return nil, fmt.Errorf("%q is missing", id)
	}

	structure, ok := element.(*Structure)
	if !ok {
		return nil, fmt.Errorf("%q is not a structure", id)
	}

	return structure, nil
}

type IdName struct {
	Id          int
	Name        string
	Description string
}

func (i *IdName) GetId() int {
	return i.Id
}
func (i *IdName) GetName() string {
	return i.Name
}
func (i *IdName) GetDescription() string {
	return i.Description
}

type Grammar struct {
	Xml *XmlGrammar

	IdName

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
	Xml *XmlStructure

	IdName

	Length       Reference
	LengthOffset Reference

	Endian   Endian
	Signed   bool
	Encoding string

	Extends Element

	//Display   Display

	Elements []Element

	/*
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
	*/

	Display      Display
	FillColour   uint32
	StrokeColour uint32
}

type GrammarRef struct {
	Xml *XmlGrammarRef

	IdName

	Filename string
	Disabled bool
}

type Custom struct {
	Xml *XmlCustom

	IdName

	Length Reference
	Script string // TODO?

	FillColour   uint32
	StrokeColour uint32
}

type StructRef struct {
	Xml *XmlStructRef

	IdName

	Structure *Structure

	RepeatMin Reference
	RepeatMax Reference

	FillColour   uint32
	StrokeColour uint32
}

type String struct {
	Xml *XmlString

	IdName

	Type string // TODO Convert to "StringType" // "zero-terminated", "fixed-length"

	Length    Reference
	Encoding  string
	MustMatch bool

	RepeatMin Reference
	RepeatMax Reference

	FillColour   uint32
	StrokeColour uint32

	Values []*FixedValue
}

type Binary struct {
	Xml *XmlBinary

	IdName

	Length     Reference
	LengthUnit string

	RepeatMin Reference
	RepeatMax Reference

	MustMatch bool
	Unused    bool
	Disabled  bool

	FillColour   uint32
	StrokeColour uint32

	Values []*FixedValue
}

type Number struct {
	Xml *XmlNumber

	IdName

	RepeatMin Reference
	RepeatMax Reference

	Type       string // TODO Convert to Type
	Length     Reference
	LengthUnit string // TODO Convert to LengthUnit "", "bit" (default "byte")

	Endian Endian
	Signed bool

	Display      Display
	FillColour   uint32
	StrokeColour uint32

	// TODO Handle the below fields:
	MustMatch       bool
	ValueExpression string

	MinVal string
	MaxVal string

	Disabled bool

	Values []*FixedValue
	Masks  []*Mask
}

type Offset struct {
	Xml *XmlOffset

	IdName

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

	FillColour   uint32
	StrokeColour uint32
}

type Mask struct {
	Xml *XmlMask

	Name  string
	Value string // The mask // TODO Change to uint64

	Values []*FixedValue
}

type FixedValues struct {
	Xml *XmlFixedValues

	Values []*FixedValue `xml:"fixedvalue,omitempty"`
}

type FixedValue struct {
	Xml *XmlFixedValue

	Name  string
	Value string
}

// TODO Reconsider the script elements, as they don't need to correct match the XML

type ScriptElement struct {
	Xml *XmlScriptElement

	IdName

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
