package cache

import (
	"fmt"
	"os"
)

// ParseFind parses the find.cache file and returns the search text.
// Format: size(4) + "10wf"(4) + data
func ParseFind(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading find.cache: %w", err)
	}

	if len(data) < 8 {
		return "", nil
	}

	// Extract any printable text after the header
	var text []byte
	for _, b := range data[8:] {
		if b >= 0x20 && b < 0x7F {
			text = append(text, b)
		}
	}
	return string(text), nil
}
