package main

import "testing"

func TestGenerate(t *testing.T) {
	g := Generator{
		dir: "testdata",
	}

	g.parsePackageDir(g.dir)

	// Run generate for each type.
	g.generate([]string{"Sample"})

}
