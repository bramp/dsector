# dsector [![Build Status](https://img.shields.io/travis/bramp/dsector.svg)](https://travis-ci.org/bramp/dsector) [![Coverage](https://img.shields.io/coveralls/bramp/dsector.svg)](https://coveralls.io/github/bramp/dsector) [![Report card](https://goreportcard.com/badge/github.com/bramp/dsector)](https://goreportcard.com/report/github.com/bramp/dsector)

<!-- [![GoDoc](https://godoc.org/github.com/bramp/dsector?status.svg)](https://godoc.org/github.com/bramp/dsector) -->

Dsector is a Go package that provides a API for parsing binary files using a predefined grammar. Think the wireshark dissector but for files instead of packets. This useful for debugging or inspecting files.

For example, using the [PNG grammar](grammars/png.grammar) on a [sample PNG file](samples/png/basi0g01.png) produces:
```bash
$ go run inspect/main.go grammars/png.grammar samples/png/basi0g01.png 

PNG Images: (1 children)
  [0] PNG File: (3 children)
    [0] Header: (1 children)
      [0] Eye catcher: 89504e470d0a1a0a (Eye catcher)
    [1] IHDR: (4 children)
      [0] Length: 13
      [1] Type: 0x49484452 (IHDR)
      [2] Data: (7 children)
        [0] Width: 32
        [1] Height: 32
        [2] Bit depth: 1
        [3] Color type: 0 (Type0)
        [4] Compression method: 0 (deflate/inflate)
        [5] Filter method: 0 (Adaptive)
        [6] Interlace method: 1 (Adam7 interlace)
      [3] CRC: 2c0677cf (4 bytes)
    [2] Chunks: (3 children)
      [0] gAMA: (4 children)
        [0] Length: 4
        [1] Type: 0x67414d41
        [2] Data: (1 children)
          [0] Gamma: 100000
        [3] CRC: 31e8965f (4 bytes)
      [1] IDAT: (4 children)
        [0] Length: 144
        [1] Type: 0x49444154
        [2] Data: (1 children)
          [0] Image Data: 789c2d8d310e... (144 bytes)
        [3] CRC: 661822f2 (4 bytes)
      [2] IEND: (4 children)
        [0] Length: 0
        [1] Type: 0x49454e44
        [2] Data: (0 children)
        [3] CRC: ae426082 (4 bytes)
```

Dsector is able to parse the same XML grammar files provided by [Synalyze It!](https://www.synalysis.net/) a commerical Hex Editor for the Mac. If you need a better, more complete tool, then please consider [Synalyze It!](https://www.synalysis.net/) or [Hexinator](https://hexinator.com/).

#TODO

* Better document the API.
* Move the code into multiple packages. Ideally the parsing/xml/structs would all be in seperate packages.
* More complete tests
* More sample files for each Grammar
* Run a fuzzing library to check we don't crash!
* Many more things
