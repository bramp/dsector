package ufwb

import "fmt"

func cantExtendFromError(child, parent Element) error {
	return &validationError{e: child, msg: fmt.Sprintf("element can't extend from %T", parent)}
}

func (g *Grammar) SetExtends(element Element) error {
	panic("This should never be called")
}

func (s *Structure) SetExtends(element Element) error {
	parent, ok := element.(*Structure)
	if !ok {
		return cantExtendFromError(s, element)
	}

	s.extends = parent
	if len(s.elements) < len(parent.elements) {
		return &validationError{e: s, msg: "child element must have atleast as many elements as the parent"}
	}

	// Update all child
	for i, child := range parent.elements {
		if e, ok := s.elements[i].(Updatable); ok {
			if err := e.SetExtends(child); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *GrammarRef) SetExtends(element Element) error{
	parent, ok := element.(*GrammarRef);
	if !ok {
		return cantExtendFromError(g, element)
	}
	g.extends = parent
	return nil
}

func (c *Custom) SetExtends(element Element)error {
	parent, ok := element.(*Custom);
	if !ok {
		return cantExtendFromError(c, element)
	}
	c.extends = parent
	return nil
}

func (s *StructRef) SetExtends(element Element)error {
	parent, ok := element.(*StructRef);
	if !ok {
		return cantExtendFromError(s, element)
	}
	s.extends = parent
	return nil
}

func (s *String) SetExtends(element Element)error {
	parent, ok := element.(*String);
	if !ok {
		return cantExtendFromError(s, element)
	}
	s.extends = parent
	return nil
}

func (b *Binary) SetExtends(element Element) (error) {
	parent, ok := element.(*Binary);
	if !ok {
		return cantExtendFromError(b, element)
	}
	b.extends = parent
	return nil
}

func (n *Number) SetExtends(element Element) error{
	parent, ok := element.(*Number);
	if !ok {
		return cantExtendFromError(n, element)
	}
	n.extends = parent
	return nil
}

func (o *Offset) SetExtends(element Element) error{
	parent, ok := element.(*Offset);
	if !ok {
		return cantExtendFromError(o, element)
	}
	o.extends = parent
	return nil
}

func (s *ScriptElement) SetExtends(element Element) error {
	parent, ok := element.(*ScriptElement);
	if !ok {
		return cantExtendFromError(s, element)
	}
	s.extends = parent
	return nil
}
