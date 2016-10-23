// Package inspect parses a file and prints its output.
package main

import (
	"bramp.net/dsector/ufwb"
	"flag"
	"fmt"
	"os"
)

var (
	// TODO Change this to be arg1 and arg2, instead of flags
	grammar = flag.String("grammar", "", "grammar filename")
	target = flag.String("target", "", "target filename")
)

func openGrammar() *ufwb.Ufwb {
	file, err := os.Open(*grammar)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open grammar %q: %s\n", grammar, err)
		os.Exit(1)
	}
	defer file.Close()

	g, errs := ufwb.ParseXmlGrammar(file)
	if len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "failed to parse grammar %q:\n", grammar)
		for i, err := range errs {
			fmt.Fprintf(os.Stderr, "[%d] %s\n", i, err.Error())
		}
		os.Exit(1)
	}

	return g
}

func decode(u *ufwb.Ufwb, f ufwb.File) {
	decoder := ufwb.NewDecoder(u, f)
	value, err := decoder.Decode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse target: %s\n", err.Error())
		os.Exit(1)
	}

	str, err := value.Format(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to format parsed output: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Print(str)
}

func parseFlags() {
	flag.Parse()

	if grammar == nil || *grammar == "" {
		fmt.Fprintf(os.Stderr, "Please specify a grammar file\n")
		os.Exit(1)
	}

	if target == nil || *target == "" {
		fmt.Fprintf(os.Stderr, "Please specify a target file\n")
		os.Exit(1)
	}

}

func main() {

	parseFlags()

	g := openGrammar()

	file, err := ufwb.OpenOSFile(*target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open target %q: %s\n", target, err.Error())
		return
	}
	defer file.Close()

	decode(g, file)
}
