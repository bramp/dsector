// Just check we can parse all the sample files, not actually checking the parsed result is correct.
package ufwb

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"testing"
)

const (
	samplesPath = "../samples"
)

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

	// Parse
	grammar, errs := ParseXmlGrammar(bytes.NewReader(in))
	if len(errs) > 0 {
		return nil, nil, fmt.Errorf("ParseXmlGrammar(%q) failed: %s", langFile, errs)
	}

	return grammar, samples, nil
}

func TestParserPng(t *testing.T) {
	// Some files are excluded because they require more than just parsing to validate
	exclude := map[string]bool{
		"xhdn0g08.png": true, // incorrect IHDR checksum
		"xdtn0g01.png": true, // missing IDAT chunk
		"xcsn0g01.png": true, // incorrect IDAT checksum
	}

	grammar, samples, err := readGrammarAndTestData("png", exclude)
	if err != nil {
		t.Fatalf("readGrammarAndTestData(..) failed: %s", err)
	}

	for _, sample := range samples {
		expectErr := strings.HasPrefix(path.Base(sample), "x")
		testFile(t, grammar, sample, expectErr)
	}
}

func TestParserCsv(t *testing.T) {
	grammar, samples, err := readGrammarAndTestData("csv", nil)
	if err != nil {
		t.Fatalf("readGrammarAndTestData(..) failed: %s", err)
	}

	for _, sample := range samples {
		testFile(t, grammar, sample, false)
	}
}

func TestParserPcap(t *testing.T) {
	grammar, samples, err := readGrammarAndTestData("pcap", nil)
	if err != nil {
		t.Fatalf("readGrammarAndTestData(..) failed: %s", err)
	}

	for _, sample := range samples {
		testFile(t, grammar, sample, false)
	}
}
