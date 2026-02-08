package parser

import (
	"fmt"
	"os"
	"strings"
)

// ParseSignature reads the familyfile.signature file and returns its content.
func ParseSignature(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading signature: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}
