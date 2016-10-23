package ufwb

import (
	"bytes"
	"io/ioutil"
	"path"
	"strings"
	"testing"
)

func TestParserPcap(t *testing.T) {
	lang := "pcap"
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

	for _, sample := range files {
		name := sample.Name()
		if !strings.HasSuffix(name, ".cap") {
			continue
		}

		// Now test
		filename := path.Join(root, sample.Name())
		testFile(t, grammar, filename, false)
	}

}
