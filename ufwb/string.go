package ufwb

// ellipsis returns the string abbreviated to length characters
func ellipsis(s string, length int) string {
	if len(s) > length {
		return s[:length - 3] + "..."
	}
	return s
}
