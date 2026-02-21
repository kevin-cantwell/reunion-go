package cache

import (
	"fmt"
	"os"

	"github.com/kedoco/reunion-explore/internal/binutil"
	"github.com/kedoco/reunion-explore/model"
)

// ParseShNames parses the shNames.cache file (searchable full names).
// Format: size(4) + count(u16) + padding(2) + "10hSan"(6) + padding(6) = 20-byte header
// Then: packed name records.
func ParseShNames(path string) ([]model.SearchName, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading shNames.cache: %w", err)
	}

	if len(data) < 20 {
		return nil, fmt.Errorf("shNames.cache too short: %d bytes", len(data))
	}

	count, _ := binutil.U16LE(data, 4)

	// Data starts after the header area. Look for name records.
	// Names appear to be length-prefixed or null-terminated after the header.
	var names []model.SearchName

	// Skip the header (approximately 20 bytes)
	pos := 20
	for len(names) < int(count) && pos < len(data) {
		// Try reading a null-terminated string
		start := pos
		for pos < len(data) && data[pos] != 0 {
			pos++
		}
		if pos > start {
			name := string(data[start:pos])
			// Filter: only include entries that look like names (printable characters)
			printable := true
			for _, c := range name {
				if c < 0x20 || c > 0x7E {
					// Allow common UTF-8 chars
					if c < 0x80 {
						printable = false
						break
					}
				}
			}
			if printable && len(name) > 0 {
				names = append(names, model.SearchName{Name: name})
			}
		}
		pos++ // skip null terminator
	}

	return names, nil
}
