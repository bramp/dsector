//
package ufwb

import (
	"encoding/xml"
	"io"
)

func ParseXmlGrammar(r io.Reader) (*Ufwb, []error) {

	// 1. Decode the xml into our XML objects
	x := &XmlUfwb{}
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(x); err != nil {
		return nil, []error{err}
	}

	// 2. Turn the XML objects into a native objects
	u, errs := x.toElement()
	if len(errs) > 0 {
		return u, errs
	}

	// 3. Now run over them, to build up indexes, relationships, etc
	return u, u.update()
}

func WriteXmlGrammar(w io.Writer, ufwb *Ufwb) error {
	w.Write([]byte(xml.Header))
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "    ")
	// TODO Instead of depending on the Xml field, recreate all the Xml, so changes in ufwb are reflected.
	return encoder.Encode(ufwb.Xml)
}
