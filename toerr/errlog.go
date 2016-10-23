// Package errors is a simple library to collect multiple errors.
package toerr

// TODO Look at https://play.golang.org/p/X2Q2aVJweW

type Errors struct {
	errs []error
}

func (e *Errors) Append(err error) {
	e.errs = append(e.errs, err)
}

func (e *Errors) Ok() bool {
	return len(e.errs) == 0
}

func (e *Errors) NotOk() bool {
	return !e.Ok()
}

func (e *Errors) Slice() []error {
	return e.errs
}

func (e *Errors) Error() string {
	// TODO Join all the errors together.
	if len(e.errs) > 0 {
		return e.errs[0].Error()
	}
	return ""
}
