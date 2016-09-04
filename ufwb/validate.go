package ufwb

import (
	"fmt"
	"reflect"
)

var (
	// Valid values for tagged fields. The first value is the default if otherwise not specified.
	// TODO Drop this table, as we correctly parse them all out
	validValues = map[string][]string{
		"bool":   []string{"no", "yes"},
		"endian": []string{"little", "big", "dynamic"},
		"lang":   []string{"Lua", "Python"},
		"lengthunit": []string{"byte", "bit"},
		"display": []string{"dec", "hex", "binary"}, // TODO Maybe "offset"? // TODO "dec" was a guess
		"string-type": []string{"zero-terminated", "fixed-length", "pascal"},
		"number-type": []string{"integer", "float"},
	}
)

type validationError struct {
	e interface{}
	msg string
}

func (e *validationError) Error() string {
	return fmt.Sprintf("<%T id=%d name=%q>: %s", e.e, getId(e.e), getName(e), e.msg)
}

type validateable interface {
	validate(u *Ufwb) error
}

func (u *Ufwb) validate() error {
	u.validateFields(u)
	u.validateFields(u.Grammar)
	return u.Grammar.validate(u)
}

func (u *Ufwb) getStructure(id string) (*Structure, error) {
	element, found := u.elements[id]
	if !found {
		return nil, fmt.Errorf("%q is missing", id)
	}

	structure, ok := element.(*Structure)
	if !ok {
		return nil, fmt.Errorf("%q is not a structure", id)
	}

	return structure, nil
}

func (g *Grammar) validate(u *Ufwb) error {
	for _, s := range g.Structures {
		if err := u.validateFields(s); err != nil {
			return err
		}

		if err := s.validate(u); err != nil {
			return err
		}
	}
	return nil
}

func (s *Structure) validate(u *Ufwb) error {
	for _, e := range s.Elements {
		if err := u.validateFields(s); err != nil {
			return err
		}

		if v, ok := e.(validateable); ok {
			if err := v.validate(u); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *StructRef) validate(u *Ufwb) error {
	structure, err := u.getStructure(s.Structure)
	s.structure = structure
	return err
}

func (n *Number) validate(u *Ufwb) error {
	if n.Length == "" {
		return &validationError{e: n, msg: "missing length field"}
	}
	return nil
}


// validateFields checks the 'ufwb' tags and validates
func (u *Ufwb) validateFields(e interface{}) error {
	s := reflect.ValueOf(e).Elem()

	// Now check all the tags on the struct
	for i := 0; i < s.NumField(); i++ {
		sf := s.Type().Field(i)
		if tag, ok := sf.Tag.Lookup("ufwb"); ok {
			f := s.Field(i)
			if f.Kind() != reflect.String || f.String() == "" {
				continue
			}

			switch tag {
			case "id":
				id := f.String()
				if _, ok := u.elements[id]; !ok {
					return &validationError{e: e, msg: fmt.Sprintf("%s refers to invalid id %q", sf.Name, id)}
				}

			case "encoding":
			// TODO Support encoding
			// Encoding:[ ANSI_X3.4-1968 IBM500 IBM850 ISCII,version=0 ISO_8859-1:1987 Shift_JIS UTF-16 UTF-16BE UTF-16LE UTF-7 UTF-8 macintosh]
			case "colour":
				colour := f.String()
				if !colourRegex.MatchString(colour) {
					return &validationError{
						e: e,
						msg: fmt.Sprintf("%s refers to invalid color %q", sf.Name, colour),
					}
				}
			case "ref":
			// TODO Support ref. Must be a number, or a reference an existing field, or be some kind of eval

			default:
				values, found := validValues[tag]
				if !found {
					// Panic because this is a programming error (not a runtime one)
					panic(fmt.Sprintf("Unknown tag value %q on %s", tag, sf.Name))
				}

				value := f.String()
				if !contains(value, values) {
					return &validationError{
						e: e,
						msg: fmt.Sprintf("%s contains invalid value %q must be one of %q", sf.Name, value, values),
					}
				}
			}
		}
	}

	return nil
}
