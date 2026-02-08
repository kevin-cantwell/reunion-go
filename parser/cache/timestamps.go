package cache

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/kevin-cantwell/reunion-go/internal/binutil"
	"github.com/kevin-cantwell/reunion-go/model"
)

// ParseTimestamps parses the timestamps.cache file.
// Format: size(4) + "icst"(4) + count(4) + extra(4) = 16-byte header
// Then: count * 20-byte fixed records.
func ParseTimestamps(path string) ([]model.TimestampEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading timestamps.cache: %w", err)
	}

	if len(data) < 16 {
		return nil, fmt.Errorf("timestamps.cache too short: %d bytes", len(data))
	}

	magic := string(data[4:8])
	if magic != "icst" {
		return nil, fmt.Errorf("timestamps.cache: unexpected magic %q", magic)
	}

	count, _ := binutil.U32LE(data, 8)

	const headerSize = 16
	const recordSize = 20

	entries := make([]model.TimestampEntry, 0, count)
	for i := uint32(0); i < count; i++ {
		off := headerSize + int(i)*recordSize
		if off+recordSize > len(data) {
			break
		}
		rec := make([]byte, recordSize)
		copy(rec, data[off:off+recordSize])
		entries = append(entries, model.TimestampEntry{
			Data: rec,
			Hex:  hex.EncodeToString(rec),
		})
	}

	return entries, nil
}
