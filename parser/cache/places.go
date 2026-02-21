package cache

import (
	"fmt"
	"os"

	"github.com/kedoco/reunion-explore/internal/binutil"
	"github.com/kedoco/reunion-explore/model"
)

// ParsePlaces parses the places.cache file.
// Format: size(4) + "ahcp"(4) + count(4) + extra(4) = 16-byte header
// Then: offset table of count * uint32
// Each record at offset: size(4) + id(4) + ref(8) + UTF-8 string
func ParsePlaces(path string) ([]model.Place, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading places.cache: %w", err)
	}

	if len(data) < 16 {
		return nil, fmt.Errorf("places.cache too short: %d bytes", len(data))
	}

	magic := string(data[4:8])
	if magic != "ahcp" {
		return nil, fmt.Errorf("places.cache: unexpected magic %q", magic)
	}

	count, _ := binutil.U32LE(data, 8)
	// Extra 4 bytes at offset 12 (skip)

	// Offset table starts at byte 16
	offsets, err := ReadOffsetTable(data, 16, int(count))
	if err != nil {
		return nil, fmt.Errorf("places.cache offset table: %w", err)
	}

	places := make([]model.Place, 0, count)
	for _, off := range offsets {
		o := int(off)
		if o+16 > len(data) {
			continue
		}
		recSize, _ := binutil.U32LE(data, o)
		id, _ := binutil.U32LE(data, o+4)
		ref := make([]byte, 8)
		copy(ref, data[o+8:o+16])

		strLen := int(recSize) - 16
		var name string
		if strLen > 0 && o+16+strLen <= len(data) {
			name = string(data[o+16 : o+16+strLen])
		}

		places = append(places, model.Place{
			ID:   id,
			Name: name,
			Ref:  ref,
		})
	}

	return places, nil
}
