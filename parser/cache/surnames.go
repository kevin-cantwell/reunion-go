package cache

import (
	"fmt"
	"os"
	"strings"

	"github.com/kevin-cantwell/reunion-go/model"
)

// ParseSurnames parses the surnames.cache file.
// Format: size(4) + "10ns"(4) + packed data
// Records are parenthesized entries like "(SURNAME, GIVEN))" with binary separators.
func ParseSurnames(path string) ([]model.SurnameEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading surnames.cache: %w", err)
	}

	if len(data) < 8 {
		return nil, fmt.Errorf("surnames.cache too short: %d bytes", len(data))
	}

	magic := string(data[4:8])
	if magic != "10ns" {
		return nil, fmt.Errorf("surnames.cache: unexpected magic %q", magic)
	}

	// Scan for parenthesized entries: (SURNAME, GIVEN))
	var entries []model.SurnameEntry
	content := string(data[8:])

	// Find all entries delimited by parentheses
	i := 0
	for i < len(content) {
		start := strings.IndexByte(content[i:], '(')
		if start == -1 {
			break
		}
		start += i
		// Find the matching close - could be )) or )
		end := strings.Index(content[start:], ")")
		if end == -1 {
			break
		}
		end += start

		entry := content[start : end+1]
		// Clean up: remove outer parens
		inner := strings.TrimPrefix(entry, "(")
		inner = strings.TrimSuffix(inner, ")")

		var se model.SurnameEntry
		se.RawEntry = entry
		parts := strings.SplitN(inner, ", ", 2)
		if len(parts) >= 1 {
			se.Surname = parts[0]
		}
		if len(parts) >= 2 {
			se.GivenName = parts[1]
		}
		entries = append(entries, se)

		// Skip past any trailing close parens
		i = end + 1
		for i < len(content) && content[i] == ')' {
			i++
		}
	}

	return entries, nil
}
