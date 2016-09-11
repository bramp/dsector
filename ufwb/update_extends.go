package ufwb

func (s *Structure) updateExtends(u *Ufwb) error {
	e, err := u.Get(s.Xml.Extends)
	if e == nil || err != nil {
		return err
	}
	if ss, ok := e.(*Structure); ok {
		s.extends = ss
		return nil
	}
	return &validationError{e: s, msg: "extend element is not a Structure"}
}


func (g *Grammar) updateExtends(u *Ufwb) error{
	// Do nothing
	return nil
}

func (g *GrammarRef) updateExtends(u *Ufwb) error{
	// Do nothing
	return nil
}

func (c *Custom) updateExtends(u *Ufwb)error {
	// Do nothing
	return nil
}

func (s *StructRef) updateExtends(u *Ufwb)error {
	// Do nothing
	return nil
}

func (s *String) updateExtends(u *Ufwb)error {
	// Do nothing
	return nil
}

func (g *Binary) updateExtends(u *Ufwb) error{
	// Do nothing
	return nil
}

func (n *Number) updateExtends(u *Ufwb) error{
	// Do nothing
	return nil
}

func (o *Offset) updateExtends(u *Ufwb) error{
	// Do nothing
	return nil
}

func (s *ScriptElement) updateExtends(u *Ufwb) error{
	// Do nothing
	return nil
}
