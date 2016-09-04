//
// Shout out to https://github.com/wicast/xj2s for helping to generate the XML structs
// TODO Consider having XML representation, and completely parsed (ie bool fields, instead of string
// fields with "yes" in them.)
package ufwb

import (
	"encoding/xml"
	"io"
)

type Formatter interface {
	// Format returns the display string for this Element
	Format(file File, value *Value) (string, error)
}

type Element interface {
	Reader
	Formatter
}

func ParseGrammar(r io.Reader) (*Ufwb, error) {
	u := &Ufwb{}

	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(u); err != nil {
		return nil, err
	}

	if err := u.build(); err != nil {
		return u, err
	}

	return u, u.validate()
}

func WriteGrammar(w io.Writer, ufwb *Ufwb) error {
	w.Write([]byte(xml.Header))
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "    ")
	return encoder.Encode(ufwb)
}
