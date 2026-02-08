package cache

import (
	"fmt"
	"os"
)

// ParseShGeneral parses the shGeneral.cache file.
// Format: size(4) + padding(4) + "10hSeg"(6) + padding
func ParseShGeneral(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading shGeneral.cache: %w", err)
	}
	return data, nil
}
