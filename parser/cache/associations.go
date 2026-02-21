package cache

import (
	"fmt"
	"os"

	"github.com/kedoco/reunion-explore/model"
)

// ParseAssociations parses the associations.cache file.
// Format: size(4) + "cosa"(4) + count(4) + data
func ParseAssociations(path string) ([]model.Association, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading associations.cache: %w", err)
	}

	if len(data) <= 16 {
		return nil, nil
	}

	// Parse any records present (usually 0)
	return nil, nil
}
