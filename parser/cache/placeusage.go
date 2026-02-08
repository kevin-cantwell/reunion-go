package cache

import (
	"fmt"
	"os"

	"github.com/kevin-cantwell/reunion-go/internal/binutil"
	"github.com/kevin-cantwell/reunion-go/model"
)

// ParsePlaceUsage parses the placeUsage.cache file.
// Format: size(4) + "hcup"(4) + count(4) + extra(4) = 16-byte header
// Then: 4-byte sub-header, followed by count variable-length records.
// Each record: total_size(4) + n_entries(4) + place_id(4) + zero(4) + [ref_id(4) + type_code(4)] * n_entries
// total_size includes the size field itself.
func ParsePlaceUsage(path string) ([]model.PlaceUsage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading placeUsage.cache: %w", err)
	}

	if len(data) < 16 {
		return nil, fmt.Errorf("placeUsage.cache too short: %d bytes", len(data))
	}

	magic := string(data[4:8])
	if magic != "hcup" {
		return nil, fmt.Errorf("placeUsage.cache: unexpected magic %q", magic)
	}

	count, _ := binutil.U32LE(data, 8)

	// Skip 16-byte header + 4-byte sub-header
	pos := 20
	usages := make([]model.PlaceUsage, 0, count)

	for i := uint32(0); i < count && pos+4 <= len(data); i++ {
		totalSize, _ := binutil.U32LE(data, pos)
		if totalSize < 16 || pos+int(totalSize) > len(data) {
			break
		}

		nEntries, _ := binutil.U32LE(data, pos+4)
		placeID, _ := binutil.U32LE(data, pos+8)
		// pos+12: zero/padding (skip)

		usage := model.PlaceUsage{PlaceID: placeID}
		for j := uint32(0); j < nEntries; j++ {
			entryOff := pos + 16 + int(j)*8
			if entryOff+8 > pos+int(totalSize) {
				break
			}
			refID, _ := binutil.U32LE(data, entryOff)
			typeCode, _ := binutil.U32LE(data, entryOff+4)
			usage.Entries = append(usage.Entries, model.PlaceUsageEntry{
				RefID:    refID,
				TypeCode: typeCode,
			})
		}
		usages = append(usages, usage)
		pos += int(totalSize)
	}

	return usages, nil
}
