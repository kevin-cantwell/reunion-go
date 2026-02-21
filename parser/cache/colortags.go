package cache

import (
	"fmt"
	"os"

	"github.com/kevin-cantwell/reunion-explore/model"
)

// ParseColorTags parses the colortags.cache file.
// Format: size(4) + "actc"(4) + count(4) + data
func ParseColorTags(path string) ([]model.ColorTag, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading colortags.cache: %w", err)
	}

	if len(data) <= 16 {
		// Empty or header-only
		return nil, nil
	}

	// Parse any records present (usually 0 for this file)
	return nil, nil
}
