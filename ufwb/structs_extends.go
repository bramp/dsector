package ufwb

import (
	"fmt"
)

func cantExtendFromError(child, parent ElementId) error {
	return &validationError{e: child, err: fmt.Errorf("element can't extend from %T", parent)}
}

func (g *Grammar) SetExtends(element Element) error {
	panic("This should never be called")
}

func (s *Structure) SetExtends(element Element) error {

	// TODO Ensure that no parent extends from a child, or any kind of loop

	extends, ok := element.(*Structure)
	if !ok {
		return cantExtendFromError(s, element)
	}

	s.extends = extends

	// Update all child to point to the right parent
	for _, element := range s.elements {
		if p := extends.ElementByName(element.Name()); p != nil {
			if err := element.SetExtends(p); err != nil {
				return err
			}
		}
	}

	return nil
}

// TODO Precompute this
func (s *Structure) Elements() []Element {

	if s.extends == nil {
		return s.elements
	}

	// Get the parent lists
	parent := s.extends.Elements()

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

func (g *GrammarRef) SetExtends(element Element) error {
	parent, ok := element.(*GrammarRef)
	if !ok {
		return cantExtendFromError(g, element)
	}
	g.extends = parent
	return nil
}

func (c *Custom) SetExtends(element Element) error {
	parent, ok := element.(*Custom)
	if !ok {
		return cantExtendFromError(c, element)
	}
	c.extends = parent
	return nil
}

func (s *StructRef) SetExtends(element Element) error {
	parent, ok := element.(*StructRef)
	if !ok {
		return cantExtendFromError(s, element)
	}
	s.extends = parent
	return nil
}

func (s *String) SetExtends(element Element) error {
	parent, ok := element.(*String)
	if !ok {
		return cantExtendFromError(s, element)
	}
	s.extends = parent
	return nil
}

func (b *Binary) SetExtends(element Element) error {
	parent, ok := element.(*Binary)
	if !ok {
		return cantExtendFromError(b, element)
	}
	b.extends = parent
	return nil
}

func (n *Number) SetExtends(element Element) error {
	parent, ok := element.(*Number)
	if !ok {
		return cantExtendFromError(n, element)
	}
	n.extends = parent
	return nil
}

func (o *Offset) SetExtends(element Element) error {
	parent, ok := element.(*Offset)
	if !ok {
		return cantExtendFromError(o, element)
	}
	o.extends = parent
	return nil
}

func (o *Script) SetExtends(element Element) error {
	parent, ok := element.(*Script)
	if !ok {
		return cantExtendFromError(o, element)
	}
	o.extends = parent
	return nil
}
