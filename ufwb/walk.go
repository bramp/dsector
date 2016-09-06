package ufwb

type WalkFunc func(root *Ufwb, element Element, parent *Structure, errs []error)

type walker struct {
	Root *Ufwb
	Func WalkFunc
}

// Walk walks the tree of Elements, starting at the bottom
func Walk(u *Ufwb, walkFunc WalkFunc) []error {
	w := walker{
		Root: u,
		Func: walkFunc,
	}
	return w.start()
}

func (walk *walker) start() []error {
	var errs []error
	walk.grammer(walk.Root.Grammar, errs)
	return errs
}

func (walk *walker) grammer(grammar *Grammar, errs []error) {
	walk.Func(walk.Root, grammar, nil, errs)

	for _, e := range grammar.Elements {
		walk.element(e, nil, errs)
	}
}

func (walk *walker) element(element Element, parent *Structure, errs []error) {
	walk.Func(walk.Root, element, parent, errs)

	if s, ok := element.(*Structure); ok {
		for _, e := range s.Elements {
			walk.element(e, s, errs)
		}
	}
}
