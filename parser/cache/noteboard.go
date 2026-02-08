package cache

import (
	"fmt"
	"os"
)

// ParseNoteboard parses the noteboard.cache file.
// Format: small binary file with magic "10bn"/"1sbn".
func ParseNoteboard(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading noteboard.cache: %w", err)
	}
	return data, nil
}
