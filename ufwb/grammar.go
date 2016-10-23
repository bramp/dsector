//
package ufwb

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"bramp.net/dsector/toerr"
)

func indexer(u *Ufwb, element Element, parent *Structure, errs *toerr.Errors) {
	_ = parent

	if _, ok := element.(*Grammar); ok {
		// Skip over Grammar elements
		return
	}

	id := element.Id()
	if id == 0 {
		errs.Append(&validationError{e: element, err: errors.New("missing id field")})
		return
	}

	// TODO Check we don't replace existing IDs
	u.Elements["id:"+strconv.Itoa(id)] = element

	if name := element.Name(); name != "" {
		u.Elements[name] = element
	}
}

func extender(u *Ufwb, element Element, parent *Structure, errs *toerr.Errors) {
	_ = parent

	// Structure is the only one with an Extends field
	s, ok := element.(*Structure)
	if !ok {
		// Skip non-structures
		return
	}

	if s.Xml.Extends != "" {
		e, ok := u.Get(s.Xml.Extends)
		if !ok {
			errs.Append(&validationError{e: element, err: fmt.Errorf("extends element %q not found", s.Xml.Extends)})
			return
		}

		if err := s.SetExtends(e); err != nil {
			errs.Append(err)
		}
	}
}

func updater(u *Ufwb, element Element, parent *Structure, errs *toerr.Errors) {
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

	// TODO add function that check if there are now any loops due to the extends and parents

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
