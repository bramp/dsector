// Inspect parses a file and prints its output.
package main

import (
	"bramp.net/dsector/input"
	"bramp.net/dsector/ufwb"
	"flag"
	"fmt"
	"os"
)

func openGrammar(grammar string) *ufwb.Ufwb {
	file, err := os.Open(grammar)
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

func decode(u *ufwb.Ufwb, f input.Input) {
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

func main() {

	flag.Parse()
	flag.Usage = func() {
		fmt.Println("inspect [grammar] [target]")
	}

	args := flag.Args()
	if len(args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	grammar := args[0]
	target := args[1]

	g := openGrammar(grammar)

	file, err := input.OpenOSFile(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open target %q: %s\n", target, err.Error())
		return
	}
	defer file.Close()

	decode(g, file)
}
