//
package ufwb

import (
	"encoding/xml"
	"io"
	"strconv"
)

func indexer(u *Ufwb, element Element, parent *Structure, errs *Errors) {
	_ = parent

	if _, ok := element.(*Grammar); ok {
		// Skip over Grammar elements
		return
	}

	id := element.GetId()
	if id == 0 {
		errs.Append(&validationError{e: element, msg: "missing id field"})
		return
	}

	// TODO Check we don't replace existing IDs
	u.Elements["id:"+strconv.Itoa(id)] = element

	if name := element.GetName(); name != "" {
		u.Elements[name] = element
	}
}

func extender(u *Ufwb, element Element, parent *Structure, errs *Errors) {
	_ = parent

	// XMLStructure is the only one with an Extends field
	s, ok := element.(*Structure)
	if !ok {
		// Skip non-structures
		return
	}

	if e := get(u, s.Xml.Extends, errs); e != nil {
		if err := s.SetExtends(e); err != nil {
			errs.Append(err)
		}
	}
}

func updater(u *Ufwb, element Element, parent *Structure, errs *Errors) {
	if parent == nil {
		parent = defaults
	}
	element.update(u, parent, errs)
}

func ParseXmlGrammar(r io.Reader) (*Ufwb, []error) {

	// 1. Decode the xml into our XML objects
	x := &XmlUfwb{}
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(x); err != nil {
		return nil, []error{err}
	}

	// 2. Turn the XML objects into a native objects
	//    This does very little sanity checking
	u, errs := x.transform()
	if len(errs) > 0 {
		return u, errs
	}

	// Building the initial id index
	if errs := Walk(u, indexer); len(errs) > 0 {
		return u, errs
	}

	// Setup the extending
	if errs := Walk(u, extender); len(errs) > 0 {
		return u, errs
	}

	// Now update and parsing all values
	if errs := Walk(u, updater); len(errs) > 0 {
		return u, errs
	}

	return u, nil
}

func WriteXmlGrammar(w io.Writer, ufwb *Ufwb) error {
	w.Write([]byte(xml.Header))
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "    ")
	// TODO Instead of depending on the Xml field, recreate all the Xml, so changes in ufwb are reflected.
	return encoder.Encode(ufwb.Xml)
}
