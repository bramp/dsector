// fetch downloads all the synalysis format files
// Shout out to https://github.com/wicast/xj2s for helping to generate the XML structs
package main // import "bramp.net/dsector/tools/fetch"

import (
	"encoding/xml"

	"io"
	"log"
	"net/http"
	"os"
	"path"
)

const indexUrl = "https://www.synalysis.net/formats.xml"
const indexLocal = "formats.xml"

type Formats struct {
	XMLName xml.Name `xml:"formats"`
	Format  []Format `xml:"format"`
}

type Format struct {
	Grammar       Grammar         `xml:"grammar"`
	Description   []Description   `xml:"description"`
	Name          Name            `xml:"name"`
	Type          Type            `xml:"type"`
	Example       Example         `xml:"example"`
	Specification []Specification `xml:"specification"`
}

type Name struct {
	Short string `xml:"short,attr"`
	Long  string `xml:"long,attr"`
}

type Grammar struct {
	Url     string `xml:"url,attr"`
	Applink string `xml:"applink,attr"`
}

type Description struct {
	Language string `xml:"language,attr"`
	Text     string `xml:",chardata"`
}

type Type struct {
	Extension string `xml:"extension,attr"`
	Uti       string `xml:"uti,attr"`
	Mimetype  string `xml:"mimetype,attr"`
}

type Example struct {
	Url string `xml:"url,attr"`
}

type Specification struct {
	Name     string `xml:"name,attr"`
	Language string `xml:"language,attr"`
	Url      string `xml:"url,attr"`
}

func fetchIndex(url, local string) {
	log.Printf("Fetching %q\n", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to fetch index %q: %s", url, err)
	}
	defer resp.Body.Close()
	// TODO Check resp is 200, and Content-Type: application/xml

	out, err := os.Create(local)
	if err != nil {
		log.Fatalf("Failed to %s", err)
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		log.Fatalf("Failed to save index: %s", err)
	}
}

func fetch(format Format) {
	log.Printf("[%s] %s: %s", format.Name.Short, format.Name.Long, format.Grammar.Url)

	resp, err := http.Get(format.Grammar.Url)
	if err != nil {
		log.Fatalf("Failed to fetch %q: %s", format.Grammar.Url, err)
	}
	defer resp.Body.Close()

	name := path.Base(format.Grammar.Url)

	out, err := os.Create(name)
	if _, err := io.Copy(out, resp.Body); err != nil {
		log.Printf("Failed to save %s: %s", name, err)
	}
}

func main() {
	fetchIndex(indexUrl, indexLocal)

	// Re-read the index
	in, err := os.Open(indexLocal)
	if err != nil {
		log.Fatalf("Failed to %s", err)
	}
	defer in.Close()

	formats := Formats{}
	decoder := xml.NewDecoder(in)
	if err := decoder.Decode(&formats); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}

	for _, format := range formats.Format {
		fetch(format)
	}
}
