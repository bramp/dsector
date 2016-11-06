package ufwb

import (
	"fmt"
)

func cantDeriveFromError(child, parent ElementId) error {
	return &validationError{e: child, err: fmt.Errorf("element can't derive from %T", parent)}
}

func (g *Grammar) DeriveFrom(element Element) error {
	panic("This should never be called")
}

func (*Padding) DeriveFrom(e Element) error {
	panic("This should never be called")
}

func (s *Structure) DeriveFrom(element Element) error {

	// TODO Ensure that no parent derives from a child, or any kind of loop

	derives, ok := element.(*Structure)
	if !ok {
		return cantDeriveFromError(s, element)
	}

	s.derives = derives

	// Update all child to point to the right parent
	for _, element := range s.elements {
		if p := derives.ElementByName(element.Name()); p != nil {
			if err := element.DeriveFrom(p); err != nil {
				return err
			}
		}
	}

	return nil
}

// TODO Precompute this
func (s *Structure) Elements() []Element {

	if s.derives == nil {
		return s.elements
	}

	// Get the parent lists
	parent := s.derives.Elements()

	// Make a copy, as to not modify the parent
	elements := make([]Element, len(parent), len(parent))
	copy(elements, parent)

	// Now for each child, merge it into the larger list
	for _, e := range s.elements {
		i, _ := Elements(elements).Find(e.Name())
		if i >= 0 {
			// Replace the existing element
			elements[i] = e
		} else {
			elements = append(elements, e)
		}
	}

	return elements
}

func (g *GrammarRef) DeriveFrom(element Element) error {
	parent, ok := element.(*GrammarRef)
	if !ok {
		return cantDeriveFromError(g, element)
	}
	g.derives = parent
	return nil
}

func (c *Custom) DeriveFrom(element Element) error {
	parent, ok := element.(*Custom)
	if !ok {
		return cantDeriveFromError(c, element)
	}
	c.derives = parent
	return nil
}

func (s *StructRef) DeriveFrom(element Element) error {
	parent, ok := element.(*StructRef)
	if !ok {
		return cantDeriveFromError(s, element)
	}
	s.derives = parent
	return nil
}

func (s *String) DeriveFrom(element Element) error {
	parent, ok := element.(*String)
	if !ok {
		return cantDeriveFromError(s, element)
	}
	s.derives = parent
	return nil
}

func (b *Binary) DeriveFrom(element Element) error {
	parent, ok := element.(*Binary)
	if !ok {
		return cantDeriveFromError(b, element)
	}
	b.derives = parent
	return nil
}

func (n *Number) DeriveFrom(element Element) error {
	parent, ok := element.(*Number)
	if !ok {
		return cantDeriveFromError(n, element)
	}
	n.derives = parent
	return nil
}

func (o *Offset) DeriveFrom(element Element) error {
	parent, ok := element.(*Offset)
	if !ok {
		return cantDeriveFromError(o, element)
	}
	o.derives = parent
	return nil
}

func (o *Script) DeriveFrom(element Element) error {
	parent, ok := element.(*Script)
	if !ok {
		return cantDeriveFromError(o, element)
	}
	o.derives = parent
	return nil
}
