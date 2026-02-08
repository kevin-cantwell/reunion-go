package cache

import (
	"fmt"
	"os"
)

// ParseDescriptions parses the descriptions.cache file.
// Format: size(4) + "idst"(4) + count(4) + extra(4) = 16 byte header
// Then: a length-prefixed string containing device+user info.
func ParseDescriptions(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading descriptions.cache: %w", err)
	}

	if len(data) <= 16 {
		return "", nil
	}

	// Extract printable text from the data section
	var text []byte
	for _, b := range data[16:] {
		if b >= 0x20 && b < 0x7F {
			text = append(text, b)
		} else if b == 0 && len(text) > 0 {
			// Null separator between strings
			text = append(text, ' ')
		}
	}
	return string(text), nil
}
