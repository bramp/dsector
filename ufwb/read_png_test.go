package ufwb

import (
	"bytes"
	"io/ioutil"
	"path"
	"strings"
	"testing"
)

func TestParserPng(t *testing.T) {
	lang := "png"
	langFile := lang + ".grammar"

	in, err := readGrammar(langFile)
	if err != nil {
		t.Fatalf("readGrammar(%q) = %s want nil error", langFile, err)
	}

	// Parse
	grammar, errs := ParseXmlGrammar(bytes.NewReader(in))
	if len(errs) > 0 {
		t.Fatalf("Parse(%q) = %q want nil error", langFile, errs)
	}

	// Now read each sample png file:
	root := path.Join(samplesPath, lang)
	files, err := ioutil.ReadDir(root)
	if err != nil {
		t.Fatalf("ioutil.ReadDir(%q) = %q want nil error", root, err)
	}

	// Some files are excluded because they require more than just parsing to validate
	exclude := []string{
		"xhdn0g08.png", // incorrect IHDR checksum
		"xdtn0g01.png", // missing IDAT chunk
		"xcsn0g01.png", // incorrect IDAT checksum
	}

	for _, sample := range files {
		name := sample.Name()
		if contains(exclude, name) || !strings.HasSuffix(name, ".png") {
			continue
		}

		// PNG starting with x are corrupt (expect failures)
		expectErr := strings.HasPrefix(name, "x")

		// Now test
		filename := path.Join(root, sample.Name())
		testFile(t, grammar, filename, expectErr)
	}

}
