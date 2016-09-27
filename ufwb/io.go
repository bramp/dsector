package ufwb

// seekUntil reads from the file until it finds a delim character, or EOF. Returning the number
// of bytes until it was found.
func seekUntil(f File, delim byte, len int64) (int64, error) {
	var n int64
	for n < len {
		b, err := f.ReadByte()
		if err != nil {
			return n, err
		}
		n++
		if b == delim {
			break
		}
	}
	return n, nil
}

/*
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
*/
