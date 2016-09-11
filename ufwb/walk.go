package ufwb

// TODO Move Errors out of this API, we shouldn't expose that, instead return a single error
type WalkFunc func(root *Ufwb, element Element, parent *Structure, errs *Errors)

type walker struct {
	Root *Ufwb
	Func WalkFunc
	errs Errors
}

// Walk walks the tree of Elements in depth-first order
func Walk(u *Ufwb, walkFunc WalkFunc) []error {
	w := walker{
		Root: u,
		Func: walkFunc,
	}
	return w.start()
}

func (walk *walker) start() []error {
	walk.grammer(walk.Root.Grammar)
	return walk.errs.Slice()
}

func (walk *walker) grammer(grammar *Grammar) {
	walk.Func(walk.Root, grammar, nil, &walk.errs)

	for _, e := range grammar.Elements {
		walk.element(e, nil)
	}
}

func (walk *walker) element(element Element, parent *Structure) {
	walk.Func(walk.Root, element, parent, &walk.errs)

	if s, ok := element.(*Structure); ok {
		for _, e := range s.elements {
			walk.element(e, s)
		}
	}
}
