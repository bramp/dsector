// Just check we can parse all the sample files, not actually checking the parsed result is correct.
package ufwb

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"path"
	"strings"
	"testing"
)

const (
	samplesPath = "../samples"
)

// readGrammarAndTestData returns the newly parsed Ufwb, and a slice of full path names to sample
// data.
func readGrammarAndTestData(lang string, exclude map[string]bool) (*Ufwb, []string, error) {
	langFile := lang + ".grammar"

	in, err := readGrammar(langFile)
	if err != nil {
		return nil, nil, err
	}

	// Read each sample file
	root := path.Join(samplesPath, lang)
	files, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, nil, fmt.Errorf("ioutil.ReadDir(%q) failed: %s", langFile, err)
	}

	samples := make([]string, 0, 0)
	for _, sample := range files {
		name := sample.Name()
		if path.Ext(name) != "."+lang {
			continue
		}

		if _, found := exclude[name]; found {
			continue
		}

		samples = append(samples, path.Join(root, name))
	}

	if len(samples) == 0 {
		return nil, nil, errors.New("no samples found")
	}

	// Parse
	grammar, errs := ParseXmlGrammar(bytes.NewReader(in))
	if len(errs) > 0 {
		return nil, nil, fmt.Errorf("ParseXmlGrammar(%q) failed: %s", langFile, errs)
	}

	return grammar, samples, nil
}

func TestParserPng(t *testing.T) {
	// Some files should fail, but require more than just parsing to validated them
	pass := map[string]bool{
		"xhdn0g08.png": true, // incorrect IHDR checksum
		"xdtn0g01.png": true, // missing IDAT chunk
		"xcsn0g01.png": true, // incorrect IDAT checksum
	}

	grammar, samples, err := readGrammarAndTestData("png", nil)
	if err != nil {
		t.Fatalf("readGrammarAndTestData(..) failed: %s", err)
	}

	samples = []string{"../samples/png/PngSuite.png"}

	for _, sample := range samples {
		expectErr := strings.HasPrefix(path.Base(sample), "x")
		if _, found := pass[path.Base(sample)]; found {
			expectErr = false
		}
		testFile(t, grammar, sample, expectErr)
	}
}

func TestParserCsv(t *testing.T) {

	grammar, samples, err := readGrammarAndTestData("csv", nil)
	if err != nil {
		t.Fatalf("readGrammarAndTestData(..) failed: %s", err)
	}

	testFile(t, grammar, "../samples/csv/empty.csv", false)
	return

	for _, sample := range samples {
		testFile(t, grammar, sample, false)
	}
}

/*
func TestParserPcap(t *testing.T) {
	grammar, samples, err := readGrammarAndTestData("pcap", nil)
	if err != nil {
		t.Fatalf("readGrammarAndTestData(..) failed: %s", err)
	}

	for _, sample := range samples {
		testFile(t, grammar, sample, false)
	}
}
*/
