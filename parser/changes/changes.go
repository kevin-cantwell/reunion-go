package changes

import (
	"fmt"
	"os"

	"github.com/kedoco/reunion-explore/model"
)

// ParseChanges parses a .changes file containing sync log records.
// Format: magic "0sfr" followed by variable-length records.
func ParseChanges(path string) ([]model.ChangeRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading changes file %s: %w", path, err)
	}

	if len(data) < 4 {
		return nil, nil
	}

	// Scan for records by looking for the record marker pattern
	var records []model.ChangeRecord
	pos := 0

	for pos < len(data) {
		// Each record has a size prefix (u32 LE)
		if pos+4 > len(data) {
			break
		}
		size := int(data[pos]) | int(data[pos+1])<<8 | int(data[pos+2])<<16 | int(data[pos+3])<<24
		if size <= 0 || pos+4+size > len(data) {
			break
		}

		rec := make([]byte, size)
		copy(rec, data[pos+4:pos+4+size])
		records = append(records, model.ChangeRecord{
			Offset: pos,
			Size:   size,
			Data:   rec,
		})
		pos += 4 + size
	}

	return records, nil
}
