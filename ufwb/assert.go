package ufwb

import "fmt"

const DEBUG = true

func assert(expr bool, format string, a ...interface{}) {
	if DEBUG && !expr {
		panic(fmt.Sprintf(format, a...))
	}
}
