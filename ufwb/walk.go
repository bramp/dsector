package ufwb

import "bramp.net/dsector/toerr"

// TODO Move Errors out of this API, we shouldn't expose that, instead return a single error
type WalkFunc func(root *Ufwb, element Element, parent *Structure, errs *toerr.Errors)

type walker struct {
	Root *Ufwb
	Func WalkFunc
	errs toerr.Errors
}

// Walk walks the tree of Elements in depth-first order
func Walk(u *Ufwb, walkFunc WalkFunc) []error {
	return WalkFrom(u, u.Grammar, walkFunc)
}

func WalkFrom(u *Ufwb, element Element, walkFunc WalkFunc) []error {
	w := walker{
		Root: u,
		Func: walkFunc,
	}
	w.element(element, nil)
	return w.errs.Slice()
}

func (walk *walker) element(element Element, parent *Structure) {
	walk.Func(walk.Root, element, parent, &walk.errs)

	if g, ok := element.(*Grammar); ok {
		for _, e := range g.Elements {
			walk.element(e, nil)
		}
	}

	if s, ok := element.(*Structure); ok {
		for _, e := range s.elements {
			walk.element(e, s)
		}
	}
}
