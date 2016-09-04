package ufwb

type WalkFunc func(root *Ufwb, element Element, parent Element) error

type walker struct {
	Root *Ufwb
	Func WalkFunc
}

// Walk walks the tree of Elements, starting at the bottom
func Walk(u *Ufwb, walkFunc WalkFunc) error {
	w := walker{
		Root: u,
		Func: walkFunc,
	}
	return w.start()
}

func (walk *walker) start() error {
	return walk.grammer(walk.Root.Grammar)
}

func (walk *walker) grammer(grammar *Grammar) error {
	for _, s := range grammar.Structures {
		if err := walk.element(s, nil); err != nil {
			return err
		}
	}

	return walk.Func(walk.Root, grammar, nil)
}

func (walk *walker) element(element Element, parent Element) error {
	if s, ok := element.(*Structure); ok {
		for _, e := range s.Elements {
			if err := walk.element(e, s); err != nil {
				return err
			}
		}
	}

	return walk.Func(walk.Root, element, parent)
}
