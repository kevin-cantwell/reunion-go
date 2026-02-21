package cache

import (
	"fmt"
	"os"

	"github.com/kevin-cantwell/reunion-explore/internal/binutil"
	"github.com/kevin-cantwell/reunion-explore/model"
)

// ParseFmNames parses the fmnames.cache file (first/given names).
// Format: size(4) + "2wps"(4) + count(4) = 12-byte header
// Then: offset table of count * uint32
// Each record at offset: size(1) + meta(5) + phonetic(2) + name_string
func ParseFmNames(path string) ([]model.FirstNameEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading fmnames.cache: %w", err)
	}

	if len(data) < 12 {
		return nil, fmt.Errorf("fmnames.cache too short: %d bytes", len(data))
	}

	// Magic is "2wps" (4 bytes), count as u32 at offset 8 includes the 'a' byte
	// which is actually 0x61 = first byte of count 353 (0x00000161 LE)
	count, _ := binutil.U32LE(data, 8)

	// Offset table starts at byte 12
	offsets, err := ReadOffsetTable(data, 12, int(count))
	if err != nil {
		return nil, fmt.Errorf("fmnames.cache offset table: %w", err)
	}

	entries := make([]model.FirstNameEntry, 0, count)
	for _, off := range offsets {
		o := int(off)
		if o >= len(data) {
			continue
		}
		recSize := int(data[o])
		if o+1+recSize > len(data) || recSize < 8 {
			continue
		}
		meta := make([]byte, 5)
		copy(meta, data[o+1:o+6])
		phonetic := string(data[o+6 : o+8])
		name := string(data[o+8 : o+1+recSize])

		entries = append(entries, model.FirstNameEntry{
			Name:     name,
			Meta:     meta,
			Phonetic: phonetic,
		})
	}

	return entries, nil
}
