package cache

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/kevin-cantwell/reunion-go/model"
)

// ParseGlobalRecords parses the globalRecords.cache file.
// Format: 20 bytes total, magic "rblg" at offset 8.
func ParseGlobalRecords(path string) (*model.GlobalRecordEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading globalRecords.cache: %w", err)
	}

	return &model.GlobalRecordEntry{
		RawData: data,
		Hex:     hex.EncodeToString(data),
	}, nil
}
