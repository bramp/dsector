package ufwb

// seekUntil reads from the file until it finds a delim character, or EOF. Returning the number
// of bytes until it was found.
func seekUntil(f File, delim byte) (int64, error) {
	n := int64(0)
	for {
		b, err := f.ReadByte()
		if err != nil {
			return n, err
		}
		n++
		if b == delim {
			return n, nil
		}
	}
}

// readBytes reads from the file until it finds a delim character, or EOF
func readBytes(f File, delim byte) ([]byte, error) {
	var line []byte

	for {
		b, err := f.ReadByte()
		if err != nil {
			return nil, err
		}
		if b == delim {
			return line, nil
		}
		line = append(line, b)
	}
}
