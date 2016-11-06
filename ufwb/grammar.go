//
package ufwb

import (
	"bramp.net/dsector/toerr"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
)

func indexer(u *Ufwb, element Element, parent *Structure, errs *toerr.Errors) {
	_ = parent

	// Handle Grammar's scripts, but otherwise skip over
	if g, ok := element.(*Grammar); ok {
		for _, s := range g.Scripts {
			if id := s.Id(); id != 0 {
				// TODO Check we don't replace existing IDs
				key := "id:" + strconv.Itoa(id)
				if _, found := u.Scripts[key]; found {
					errs.Append(&validationError{e: s, err: errors.New("duplicate id field")})
				}
				u.Scripts[key] = s
			} else {
				errs.Append(&validationError{e: s, err: errors.New("missing id field")})
			}
		}

		return
	}

	if id := element.Id(); id != 0 {
		// TODO Check we don't replace existing IDs
		key := "id:" + strconv.Itoa(id)
		if _, found := u.Elements[key]; found {
			errs.Append(&validationError{e: element, err: fmt.Errorf("%q already exists in index", key)})
		}
		u.Elements[key] = element
	}

	// Only index top level Structures by name
	if _, ok := element.(*Structure); !ok || parent != nil {
		return
	}

	if name := element.Name(); name != "" {
		if _, found := u.Elements[name]; found {
			errs.Append(&validationError{e: element, err: fmt.Errorf("%q already exists in index", name)})
		}
		u.Elements[name] = element
	}
}

// extender finds all structures, and ensures all their children extend from the correct elements.
func extender(u *Ufwb, element Element, parent *Structure, errs *toerr.Errors) {
	// Structure is the only XML element with an explicit extends field
	s, ok := element.(*Structure)
	if !ok {
		// Skip non-structures
		return
	}

	if s.Xml.Extends != "" {
		if parent != nil {
			errs.Append(&validationError{e: element, err: fmt.Errorf("only top level structs are allowed to extend: %q", s.Xml.Extends)})
			return
		}

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
	if e, ok := element.(Updatable); ok {
		e.update(u, parent, errs)
	}
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

	// 3. Building the initial id index
	if errs := Walk(u, indexer); len(errs) > 0 {
		return u, errs
	}

	// 4. Setup the extending
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
